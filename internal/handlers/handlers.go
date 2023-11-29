package handlers

import (
	"encoding/json"
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var client *plaid.APIClient

func getPlaidClient() *plaid.APIClient {
	if client == nil {
		clientOptions := plaid.NewConfiguration()
		clientOptions.AddDefaultHeader("PLAID-CLIENT-ID", os.Getenv("PLAID_CLIENT_ID"))
		clientOptions.AddDefaultHeader("PLAID-SECRET", os.Getenv("PLAID_SECRET"))

		// Use plaid.Development or plaid.Production depending on your environment
		clientOptions.UseEnvironment(plaid.Sandbox)
		client = plaid.NewAPIClient(clientOptions)
	}
	return client
}

func GetCompare(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	response, err := getMonthlyBudgetInsights()
	if err != nil {
		return err
	}

	return utils.WriteJSON(w, response)
}

func GetExpensesSummary(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	expenses, err := fetchExpenses(db2.GetDB(), firstOfMonth, lastOfMonth)
	if err != nil {
		return err
	}

	return utils.WriteJSON(w, expenses)
}

func PostForecast(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var forecast []types.Forecast
	if err := json.NewDecoder(r.Body).Decode(&forecast); err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error converting incoming json: %v", err))
	}

	if err := insertForecast(db2.GetDB(), forecast); err != nil {
		return err
	}

	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func PostExpense(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var expenses []types.Expense
	if err := json.NewDecoder(r.Body).Decode(&expenses); err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error decoding incoming expense data: %v", err))
	}

	if err := insertExpenses(db2.GetDB(), expenses); err != nil {
		return err
	}

	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func LinkBank(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	client = getPlaidClient()

	// Specify the user
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: os.Getenv("USER_ID"),
	} // Replace with the actual user ID

	// Specify the configuration for the Link token
	request := plaid.NewLinkTokenCreateRequest("XAT", "en", []plaid.CountryCode{plaid.COUNTRYCODE_US}, user)
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_AUTH, plaid.PRODUCTS_TRANSACTIONS})
	//request.SetWebhook("https://webhook-uri.com")
	//request.SetRedirectUri("https://domainname.com/oauth-page.html")
	request.SetAccountFilters(plaid.LinkTokenAccountFilters{
		Depository: &plaid.DepositoryFilter{
			AccountSubtypes: []plaid.AccountSubtype{
				plaid.ACCOUNTSUBTYPE_CHECKING,
				plaid.ACCOUNTSUBTYPE_SAVINGS,
			},
		},
	})

	// Create the Link token
	resp, _, err := client.PlaidApi.LinkTokenCreate(r.Context()).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error generating plaid client: %v", err))
	}
	linkToken := resp.GetLinkToken()
	// Print the Link token
	fmt.Println("Link token:", linkToken)
	data := map[string]interface{}{
		"LinkToken": linkToken,
	}
	tmpl := template.Must(template.ParseFiles("internal/templates/link_bank.html"))
	return tmpl.Execute(w, data)
}

func CreateItem(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	if err := r.ParseForm(); err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error parsing incoming form: %v", err))
	}
	client = getPlaidClient()
	publicToken := r.FormValue("public_token")
	errorMessage := r.FormValue("error_message")

	if errorMessage != "" {
		return utils.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Error fetching public token from link: %v", errorMessage))
	}

	exchangePublicTokenReq := plaid.NewItemPublicTokenExchangeRequest(publicToken)
	exchangedToken, _, err := client.PlaidApi.ItemPublicTokenExchange(r.Context()).ItemPublicTokenExchangeRequest(*exchangePublicTokenReq).Execute()
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error exchanging public token for access token: %v", err))
	}

	accessToken := exchangedToken.GetAccessToken()
	return utils.WriteJSON(w, accessToken)
}

func UpdateExpenseData(w http.ResponseWriter, r *http.Request) error {
	fetchedTransactions, err := retrieveTransactions(r)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error fetching transaction data: %v", err))
	}
	dbReadyExpenses, err := formatTransactionsToExpenseType(fetchedTransactions)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	}
	err = insertExpenses(db2.GetDB(), dbReadyExpenses)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	}
	success := map[string]string{
		"status": "success",
	}
	return utils.WriteJSON(w, success)
}

func RetrieveTransactions(w http.ResponseWriter, r *http.Request) error {

	client = getPlaidClient()
	log.Printf("access token used for req: %v", os.Getenv("PLAID_ACCESS_TOKEN"))
	const dateFormat = "2006-01-02"
	currentMo := time.Now()
	startDate := time.Date(currentMo.Year(), currentMo.Month(), 1, 0, 0, 0, 0, currentMo.Location()).Format(dateFormat)
	endDate := time.Now().Format(dateFormat)
	isTrue := true
	request := plaid.NewTransactionsGetRequest(os.Getenv("PLAID_ACCESS_TOKEN"), startDate, endDate)
	options := plaid.TransactionsGetRequestOptions{
		IncludePersonalFinanceCategoryBeta: &isTrue,
		Offset:                             plaid.PtrInt32(0),
		Count:                              plaid.PtrInt32(100),
	}
	request.SetOptions(options)
	transaction, _, err := client.PlaidApi.TransactionsGet(r.Context()).TransactionsGetRequest(*request).Execute()
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error requesting transctions from plaid: %v", err))
	}
	trans := transaction.Transactions
	for _, action := range trans {
		log.Printf("category: %v, name: %v, date: %v, amount: %v", action.Category, action.Name, action.Date, action.Amount)
	}
	resp := map[string]string{
		"status": "success",
	}
	return utils.WriteJSON(w, resp)
}

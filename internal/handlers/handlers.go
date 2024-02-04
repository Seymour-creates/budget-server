package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
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

// func GetCompare(w http.ResponseWriter, r *http.Request) error {
// 	if r.Method != http.MethodGet {
// 		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
// 	}

// 	response, err := getMonthlyBudgetInsights()
// 	if err != nil {
// 		return err
// 	}

// 	return utils.WriteJSON(w, response)
// }

func GetExpensesSummary(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	_ = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	// _ := firstOfMonth.AddDate(0, 1, -1)

	// expenses, err := fetchExpenses(db2.GetDB(), firstOfMonth, lastOfMonth)
	// if err != nil {
	// 	return err
	// }

	// return utils.WriteJSON(w, expenses)
	return nil
}

func PostForecast(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var forecast []types.Forecast
	if err := json.NewDecoder(r.Body).Decode(&forecast); err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error converting incoming json: %v", err))
	}

	// if err := insertForecast(db2.GetDB(), forecast); err != nil {
	// 	return err
	// }

	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func GetRight(w http.ResponseWriter, r *http.Request) error {
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

	// if err := insertExpenses(db2.GetDB(), expenses); err != nil {
	// 	return err
	// }

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
	request.SetAccountFilters(plaid.LinkTokenAccountFilters{
		Depository: &plaid.DepositoryFilter{
			AccountSubtypes: []plaid.AccountSubtype{
				plaid.ACCOUNTSUBTYPE_CHECKING,
				plaid.ACCOUNTSUBTYPE_SAVINGS,
			},
		},
	})
	request.SetRedirectUri("http://localhost:3000/assets/oauth-after-party.html")

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
	_, err = formatTransactionsToExpenseType(fetchedTransactions)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	}
	// err = insertExpenses(db2.GetDB(), dbReadyExpenses)
	// if err != nil {
	// 	return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	// }
	success := map[string]string{
		"status": "success",
	}
	return utils.WriteJSON(w, success)
}

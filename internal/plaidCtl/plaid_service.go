package plaidCtl

import (
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Service struct {
	Client *plaid.APIClient
}

func NewService(client *plaid.APIClient) *Service {
	return &Service{
		Client: client,
	}
}

func (s *Service) RetrieveTransactions(r *http.Request) ([]plaid.Transaction, *types.HTTPError) {
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
	getTransactionData, _, err := s.Client.PlaidApi.TransactionsGet(r.Context()).TransactionsGetRequest(*request).Execute()
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error requesting transctions from plaidCtl: %v", err))
	}
	transactions := getTransactionData.Transactions
	for _, action := range transactions {
		log.Printf("category: %v, name: %v, date: %v, amount: %v", action.Category, action.Name, action.Date, action.Amount)
	}
	return transactions, nil
}

func (s *Service) LinkBank(r *http.Request) (string, error) {
	client := s.Client

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
	request.SetRedirectUri(os.Getenv("LOCAL_URL") + "/assets/oauth-after-party.html")

	// Create the Link token
	resp, _, err := client.PlaidApi.LinkTokenCreate(r.Context()).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return "", utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error generating plaidCtl client: %v", err))
	}
	linkToken := resp.GetLinkToken()
	return linkToken, nil
	// Print the Link token
	//fmt.Println("Link token:", linkToken)
	//data := map[string]interface{}{
	//	"LinkToken": linkToken,
	//}
	//tmpl := template.Must(template.ParseFiles("internal/templates/link_bank.html"))
	//return tmpl.Execute(w, data)
}

func (s *Service) CreateItem(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	if err := r.ParseForm(); err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error parsing incoming form: %v", err))
	}
	client := s.Client
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
	err = os.Setenv("PLAID_ACCESS_TOKEN", accessToken)
	if err != nil {
		log.Printf("Unable to update access token in env file: %v", err)
	}
	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func getBudgetCategory(plaidCategory string) string {
	categoryMappings := map[string]string{
		"INCOME":                    "saving",
		"TRANSFER":                  "bill",
		"LOAN":                      "debt",
		"BANK FEES":                 "bill",
		"ENTERTAINMENT":             "ent",
		"FOOD AND DRINK":            "takeout",
		"GENERAL MERCHANDISE":       "misc",
		"HOME IMPROVEMENT":          "bill",
		"MEDICAL":                   "bill",
		"PERSONAL CARE":             "misc",
		"GENERAL SERVICES":          "bill",
		"GOVERNMENT AND NON PROFIT": "bill",
		"TRANSPORTATION":            "bill",
		"TRAVEL":                    "ent",
		"RENT AND UTILITIES":        "bill",
	}

	plaidCategory = strings.ToUpper(plaidCategory)
	for keyword, category := range categoryMappings {
		if strings.Contains(plaidCategory, keyword) {
			return category
		}
	}
	return "misc"
}

func cPlaidCategoryToExpense(category []string) string {
	if len(category) == 0 {
		return "misc"
	}
	for _, cat := range category {
		budgetCategory := getBudgetCategory(cat)
		if budgetCategory != "misc" {
			return budgetCategory
		}
	}
	return "misc"
}

func (s *Service) FormatTransactionsToExpenseType(transactions []plaid.Transaction) ([]types.Expense, *types.HTTPError) {
	var expenses []types.Expense
	for _, action := range transactions {
		log.Printf("category: %v, name: %v, date: %v, big amount: %v", action.Category, action.Name, action.Date, action.Amount)
		date, err := time.Parse("2006-01-02", action.Date)
		// need to find way to intelligently sort expenses into appropriate categories
		if err != nil {
			return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error parsing date: %v", err))
		}
		category := cPlaidCategoryToExpense(action.Category)
		expenses = append(expenses, types.Expense{Description: action.Name, Date: date, Category: category, Amount: float64(action.Amount)})
	}
	return expenses, nil
}
package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/plaidCtl"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
)

var client *plaid.APIClient

type Handler struct {
	plaid *plaidCtl.Service
	db    db.Repository
}

func MakeNewHttpHandler(plaidClient *plaidCtl.Service, repo db.Repository) *Handler {
	return &Handler{plaid: plaidClient, db: repo}
}

func (h *Handler) GetCompare(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	response, err := h.db.GetMonthlyBudgetInsights()
	if err != nil {
		return err
	}

	return utils.WriteJSON(w, response)
}

func (h *Handler) GetExpensesSummary(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	expenses, err := h.db.FetchExpenses(firstOfMonth, lastOfMonth)
	if err != nil {
		return err
	}

	return utils.WriteJSON(w, expenses)
}

func (h *Handler) PostForecast(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var forecast []types.Forecast
	if err := json.NewDecoder(r.Body).Decode(&forecast); err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error converting incoming json: %v", err))
	}

	if err := h.db.InsertForecast(forecast); err != nil {
		return err
	}

	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func (h *Handler) GetRight(w http.ResponseWriter, r *http.Request) error {
	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func (h *Handler) PostExpense(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	var expenses []types.Expense
	if err := json.NewDecoder(r.Body).Decode(&expenses); err != nil {
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error decoding incoming expense data: %v", err))
	}

	if err := h.db.InsertExpenses(expenses); err != nil {
		return err
	}

	return utils.WriteJSON(w, map[string]string{"status": "success"})
}

func (h *Handler) LinkBank(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	client = h.plaid.Client

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
		return utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error generating plaidCtl client: %v", err))
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

func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	if err := r.ParseForm(); err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error parsing incoming form: %v", err))
	}
	client = h.plaid.Client
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

func (h *Handler) UpdateExpenseData(w http.ResponseWriter, r *http.Request) error {
	fetchedTransactions, err := h.plaid.RetrieveTransactions(r)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error fetching transaction data: %v", err))
	}
	dbReadyExpenses, err := h.plaid.FormatTransactionsToExpenseType(fetchedTransactions)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	}
	err = h.db.InsertExpenses(dbReadyExpenses)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, err.Message)
	}
	success := map[string]string{
		"status": "success",
	}
	return utils.WriteJSON(w, success)
}

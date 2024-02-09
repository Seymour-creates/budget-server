package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/plaidCtl"
	"html/template"
	"net/http"
	"time"

	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
)

type Handler struct {
	plaid *plaidCtl.Service
	db    db.Repository
}

// MakeNewHttpHandler returns instance of Handler struct.
func MakeNewHttpHandler(plaidClient *plaidCtl.Service, repo db.Repository) *Handler {
	return &Handler{plaid: plaidClient, db: repo}
}

// GetForecastAndExpenses returns types.MonthlyBudgetInsights struct. ({ []types.Forecast, []types.Expense })
func (h *Handler) GetForecastAndExpenses(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method Not Allowed")
	}

	response, err := h.db.GetMonthlyBudgetInsights()
	if err != nil {
		return err
	}

	return utils.WriteJSON(w, response)
}

// GetExpensesSummary Returns []types.Expense for the month from db.
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

// PostForecast Post CLI user input of types.Forecast into db.
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

// PostExpense Post CLI user input of types.Expense in to db.
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

// LinkBank returns HTMX page to register users bank using Plaid Link
func (h *Handler) LinkBank(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}

	linkToken, err := h.plaid.LinkBank(r)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error retrieving link token: %v", err))
	}
	// Print the Link token
	fmt.Println("Link token:", linkToken)
	data := map[string]interface{}{
		"LinkToken": linkToken,
	}
	tmpl := template.Must(template.ParseFiles("internal/templates/link_bank.html"))
	return tmpl.Execute(w, data)
}

// CreatePlaidBankItem links users bank to APP in plaid - returns token for plaid client
func (h *Handler) CreatePlaidBankItem(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return utils.NewHTTPError(http.StatusMethodNotAllowed, "Method not allowed.")
	}
	if err := r.ParseForm(); err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error parsing incoming form: %v", err))
	}
	client := h.plaid.Client
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

// UpdateExpenseData retrieves bank transaction data from plaid & posts to db - responds with success
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

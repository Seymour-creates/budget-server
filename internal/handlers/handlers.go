package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/plaidCtl"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
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
	fromDate, toDate, _ := decodeDateRange(r)
	expenses, err := h.db.FetchExpenses(fromDate, toDate)

	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error retreiving expense data from db: %v", err))
	}
	return utils.WriteJSON(w, expenses)
}

func decodeDateRange(r *http.Request) (string, string, error) {
	var dateRange types.DateRange
	err := json.NewDecoder(r.Body).Decode(&dateRange)
	if err != nil {
		if err != io.EOF {
			// Log the error or handle it as needed
			return "", "", utils.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("error decoding json in request GetExpensesSummary(): %v", err))
		}
		// If err is io.EOF, consider it as empty body and proceed with defaults
		err = nil
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	lastOfMonth := time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 999999999, now.Location()).Format("2006-01-02")

	// Use provided dates or defaults if empty
	fromDate := firstOfMonth
	if dateRange.FromDate != "" {
		fromDate = dateRange.FromDate
	}

	toDate := lastOfMonth
	if dateRange.ToDate != "" {
		toDate = dateRange.ToDate
	}

	return fromDate, toDate, nil
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
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error posting forecast data in db: %v", err))
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

	data := map[string]interface{}{
		"LinkToken": linkToken,
		"APP_URL":   os.Getenv("APP_URL"),
	}
	tmpl := template.Must(template.ParseFiles("internal/templates/link_bank.html"))
	return tmpl.Execute(w, data)
}

func (h *Handler) OauthRedirect(w http.ResponseWriter, r *http.Request) error {
	publicToken := r.FormValue("public_token")
	errorMsg := r.FormValue("error_message")
	if errorMsg != "" {
		return utils.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Error fetching public token from link: %v", errorMsg))
	}
	data := map[string]interface{}{
		"PublicToken": publicToken,
	}
	tmpl := template.Must(template.ParseFiles("internal/templates/oauth-after.html"))
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
	publicToken := r.FormValue("public_token")
	errorMessage := r.FormValue("error_message")

	if errorMessage != "" {
		return utils.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Error fetching public token from link: %v", errorMessage))
	}

	accessToken, err := h.plaid.CreateItem(publicToken, r)
	if err != nil {
		return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error exchanging public token for access token: %v", err))
	}

	log.Printf("#*#*#*# accessToken: %v", accessToken)

	return utils.WriteJSON(w, map[string]string{"status": "success"})
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

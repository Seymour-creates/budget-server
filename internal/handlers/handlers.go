package handlers

import (
	"encoding/json"
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"net/http"
	"time"
)

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

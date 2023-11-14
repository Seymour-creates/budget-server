package handlers

import (
	"encoding/json"
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/db"
	"github.com/Seymour-creates/budget-server/types"
	"github.com/Seymour-creates/budget-server/utils"
	"net/http"
	"time"
)

func GetCompare(w http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return fmt.Errorf("method not allowed")
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// Fetch expenses
	query := "SELECT categoryID, amount, date, description FROM expenses WHERE date >= ? AND date <= ?"
	rows, err := db.Query(query, firstOfMonth, lastOfMonth)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error fetching expenses: %v", err))
		return err
	}
	defer rows.Close()

	var expenses []types.Expense
	for rows.Next() {
		var exp types.Expense
		var date string
		if err := rows.Scan(&exp.Category, &exp.Amount, &date, &exp.Description); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error scanning expenses: %v", err))
			return err
		}
		var layout = "2006-01-02" // adjust the layout to match your date format
		exp.Date, err = time.Parse(layout, date)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		expenses = append(expenses, exp)
	}
	if err = rows.Err(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error iterating expenses rows: %v", err))
		return err
	}

	// Fetch forecast
	// Adjust the query and Scan method according to your 'forecast' table structure
	// This is just an example assuming your forecast table structure.
	forecastQuery := "SELECT categoryID, amount FROM forecast WHERE period = ?"
	forecastRows, err := db.Query(forecastQuery, firstOfMonth.Format("2006-01-02"))
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error fetching forecast: %v", err))
		return err
	}
	defer forecastRows.Close()
	var forecast []types.Forecast
	for forecastRows.Next() {
		var fcast types.Forecast
		// Adjust the fields in Scan method according to your 'forecast' table
		if err := forecastRows.Scan(&fcast.Category, &fcast.Amount); err != nil {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error scanning forecast: %v", err))
			return err
		}
		forecast = append(forecast, fcast)
	}
	if err = forecastRows.Err(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error iterating forecast rows: %v", err))
		return err
	}

	response := types.MonthlyBudgetInsights{
		Expenses: expenses,
		Forecast: forecast,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("error encoding response: %v", err))
		return fmt.Errorf("error encoding response: %v", err)
	}

	return nil
}

func PostExpense(w http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return fmt.Errorf("method not allowed")
	}

	var expenses []types.Expense
	err := json.NewDecoder(r.Body).Decode(&expenses)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return err
	}

	for _, expense := range expenses {
		_, err := db.Exec(
			"INSERT INTO expenses (date, description, amount, categoryID) VALUES (?, ?, ?, ?)",
			expense.Date,
			expense.Description,
			expense.Amount,
			expense.Category,
		)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return err
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "success"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusConflict, err.Error())
		return fmt.Errorf("error formatting response: %v", err)
	}
	return nil
}

func GetExpensesSummary(w http.ResponseWriter, r *http.Request) error {
	db := db2.GetDB()
	if r.Method != http.MethodGet {
		return fmt.Errorf("method not allowed")
	}

	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	query := "SELECT categoryID, amount, date, description FROM expenses WHERE date >= ? AND date <= ?"
	rows, err := db.Query(query, firstOfMonth, lastOfMonth)
	if err != nil {
		return fmt.Errorf("error getting expenses from db: %v", err)
	}
	defer rows.Close()

	var expenses []types.Expense
	for rows.Next() {
		var date string
		var exp types.Expense
		if err := rows.Scan(&exp.Category, &exp.Amount, &date, &exp.Description); err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		var layout = "2006-01-02" // adjust the layout to match your date format
		exp.Date, err = time.Parse(layout, date)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		expenses = append(expenses, exp)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(expenses); err != nil {
		return fmt.Errorf("error encoding response: %v", err)
	}

	return nil
}

func PostForecast(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed")
	}
	db := db2.GetDB()
	var forecast []types.Forecast
	err := json.NewDecoder(r.Body).Decode(&forecast)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return fmt.Errorf("error converting body from json: %v", err)
	}

	for _, categoryForecast := range forecast {
		_, err := db.Exec("INSERT INTO forecast (categoryID, amount) VALUES (?, ?)",
			categoryForecast.Category,
			categoryForecast.Amount)
		if err != nil {
			fmt.Errorf("error posting forecast data to db: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "success"}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.WriteError(w, http.StatusConflict, err.Error())
		return fmt.Errorf("error formatting response: %v", err)
	}

	return nil
}

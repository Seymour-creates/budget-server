package handlers

import (
	"database/sql"
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/db"
	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	"github.com/plaid/plaid-go/plaid"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func fetchExpenses(db *sql.DB, start, end time.Time) ([]types.Expense, *types.HTTPError) {
	const query = `SELECT categoryID, amount, date, description FROM expenses WHERE date >= ? AND date <= ?`
	rows, err := db.Query(query, start, end)
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error fetching expenses: %v", err))
	}
	defer rows.Close()

	var expenses []types.Expense
	for rows.Next() {
		var exp types.Expense
		var date string
		if err := rows.Scan(&exp.Category, &exp.Amount, &date, &exp.Description); err != nil {
			return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error scanning expense: %v", err))
		}
		exp.Date, err = time.Parse("2006-01-02", date)
		if err != nil {
			return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error parsing expense date: %v", err))
		}
		expenses = append(expenses, exp)
	}
	if err = rows.Err(); err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error iterating expenses rows: %v", err))
	}

	return expenses, nil
}

// fetchForecast retrieves forecast data from the database for a specific period.
func fetchForecast(db *sql.DB, period time.Time) ([]types.Forecast, *types.HTTPError) {
	const forecastQuery = `SELECT categoryID, amount FROM forecast WHERE period = ?`
	forecastRows, err := db.Query(forecastQuery, period.Format("2006-01-02"))
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error fetching forecast: %v", err))
	}
	defer forecastRows.Close()

	var forecast []types.Forecast
	for forecastRows.Next() {
		var fcast types.Forecast
		if err := forecastRows.Scan(&fcast.Category, &fcast.Amount); err != nil {
			return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error scanning forecast: %v", err))
		}
		forecast = append(forecast, fcast)
	}
	if err = forecastRows.Err(); err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error iterating forecast rows: %v", err))
	}

	return forecast, nil
}

func getMonthlyBudgetInsights() (*types.MonthlyBudgetInsights, *types.HTTPError) {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	expenses, err := fetchExpenses(db.GetDB(), firstOfMonth, lastOfMonth)
	if err != nil {
		return nil, err
	}

	forecast, err := fetchForecast(db.GetDB(), firstOfMonth)
	if err != nil {
		return nil, err
	}

	return &types.MonthlyBudgetInsights{
		Expenses: expenses,
		Forecast: forecast,
	}, nil
}

func insertExpenses(db *sql.DB, expenses []types.Expense) *types.HTTPError {
	const insertQuery = "INSERT INTO expenses (date, description, amount, categoryID) VALUES (?, ?, ?, ?)"
	for _, expense := range expenses {
		_, err := db.Exec(insertQuery, expense.Date, expense.Description, expense.Amount, expense.Category)
		if err != nil {
			return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error inserting data into expenses table: %v", err))
		}
	}
	return nil
}

func insertForecast(db *sql.DB, forecast []types.Forecast) *types.HTTPError {
	const insertQuery = "INSERT INTO forecast (categoryID, amount) VALUES (?, ?)"
	for _, f := range forecast {
		_, err := db.Exec(insertQuery, f.Category, f.Amount)
		if err != nil {
			return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error posting forecast data to db: %v", err))
		}
	}
	return nil
}

func formatTransactionsToExpenseType(transactions []plaid.Transaction) ([]types.Expense, *types.HTTPError) {
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

func retrieveTransactions(r *http.Request) ([]plaid.Transaction, *types.HTTPError) {
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
	getTransactionData, _, err := client.PlaidApi.TransactionsGet(r.Context()).TransactionsGetRequest(*request).Execute()
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Error requesting transctions from plaid: %v", err))
	}
	transactions := getTransactionData.Transactions
	for _, action := range transactions {
		log.Printf("category: %v, name: %v, date: %v, amount: %v", action.Category, action.Name, action.Date, action.Amount)
	}
	return transactions, nil
}

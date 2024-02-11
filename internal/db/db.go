package db

import (
	"database/sql"
	"fmt"
	"github.com/Seymour-creates/budget-server/internal/types"
	"github.com/Seymour-creates/budget-server/internal/utils"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

type Manager struct {
	db *sql.DB
}

func NewDBManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

func (man *Manager) FetchExpenses(start, end time.Time) ([]types.Expense, *types.HTTPError) {
	const query = `SELECT categoryID, amount, date, description FROM expenses WHERE date >= ? AND date <= ?`
	rows, err := man.db.Query(query, start, end)
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error fetching expenses: %v", err))
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("error closing row: %v", err)
		}
	}(rows)

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

func (man *Manager) FetchForecast(period time.Time) ([]types.Forecast, *types.HTTPError) {
	const forecastQuery = `SELECT categoryID, amount FROM forecast WHERE period = ?`
	forecastRows, err := man.db.Query(forecastQuery, period.Format("2006-01-02"))
	if err != nil {
		return nil, utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error fetching forecast: %v", err))
	}
	defer func(forecastRows *sql.Rows) {
		err := forecastRows.Close()
		if err != nil {
			log.Printf("error closing row: %v", err)
		}
	}(forecastRows)

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

func (man *Manager) GetMonthlyBudgetInsights() (*types.MonthlyBudgetInsights, *types.HTTPError) {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	expenses, err := man.FetchExpenses(firstOfMonth, lastOfMonth)
	if err != nil {
		return nil, err
	}

	forecast, err := man.FetchForecast(firstOfMonth)
	if err != nil {
		return nil, err
	}

	return &types.MonthlyBudgetInsights{
		Expenses: expenses,
		Forecast: forecast,
	}, nil
}

func (man *Manager) InsertExpenses(expenses []types.Expense) *types.HTTPError {
	const insertQuery = "INSERT INTO expenses (date, description, amount, categoryID) VALUES (?, ?, ?, ?)"
	for _, expense := range expenses {
		_, err := man.db.Exec(insertQuery, expense.Date, expense.Description, expense.Amount, expense.Category)
		if err != nil {
			return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error inserting data into expenses table: %v", err))
		}
	}
	return nil
}

func (man *Manager) InsertForecast(forecast []types.Forecast) *types.HTTPError {
	const insertQuery = "INSERT INTO forecast (categoryID, amount) VALUES (?, ?)"
	for _, f := range forecast {
		_, err := man.db.Exec(insertQuery, f.Category, f.Amount)
		if err != nil {
			return utils.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error posting forecast data to db: %v", err))
		}
	}
	return nil
}

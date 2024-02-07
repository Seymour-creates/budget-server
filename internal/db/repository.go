package db

import (
	"github.com/Seymour-creates/budget-server/internal/types"
	"time"
)

type Repository interface {
	fetchExpenses(start, end time.Time) ([]types.Expense, error)
	fetchForecast(period time.Time) ([]types.Forecast, *types.HTTPError)
	getMonthlyBudgetInsights() (*types.MonthlyBudgetInsights, *types.HTTPError)
	insertExpenses(expenses []types.Expense) *types.HTTPError
	insertForecast(forecast []types.Forecast) *types.HTTPError
}

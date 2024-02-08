package db

import (
	"github.com/Seymour-creates/budget-server/internal/types"
	"time"
)

type Repository interface {
	FetchExpenses(start, end time.Time) ([]types.Expense, *types.HTTPError)
	FetchForecast(period time.Time) ([]types.Forecast, *types.HTTPError)
	GetMonthlyBudgetInsights() (*types.MonthlyBudgetInsights, *types.HTTPError)
	InsertExpenses(expenses []types.Expense) *types.HTTPError
	InsertForecast(forecast []types.Forecast) *types.HTTPError
}

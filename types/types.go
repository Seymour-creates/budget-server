package types

import (
	"net/http"
	"time"
)

type Expense struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
}

type Forecast struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

type MonthlyBudgetInsights struct {
	Expenses []Expense  `json:"expenses"`
	Forecast []Forecast `json:"forecast"`
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

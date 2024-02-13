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

type HTTPError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return e.Message
}

type DateRange struct {
	FromDate string `json:"fromDate"`
	ToDate   string `json:"toDate"`
}

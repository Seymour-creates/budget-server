package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Seymour-creates/budget-server/types"
	"github.com/Seymour-creates/budget-server/utils"
	"log"
	"net/http"
)

func AddExpense(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method not allowed")
	}
	_, err := fmt.Fprintf(w, "Expense added")
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}

func Compare(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return fmt.Errorf("method not allowed")
	}
	_, err := fmt.Fprintf(w, "Comparison handled")
	if err != nil {
		log.Fatal(err.Error())
	}
	return nil
}

func uploadExpensesHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}

	var expenses []types.Expense
	err := json.NewDecoder(r.Body).Decode(&expenses)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
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
			return
		}
	}

	w.WriteHeader(http.StatusOK)
} // ... other handlers

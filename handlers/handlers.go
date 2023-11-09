package handlers

import (
	"encoding/json"
	"fmt"
	db2 "github.com/Seymour-creates/budget-server/db"
	"github.com/Seymour-creates/budget-server/types"
	"github.com/Seymour-creates/budget-server/utils"
	"log"
	"net/http"
)

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

func AddExpense(w http.ResponseWriter, r *http.Request) error {
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
	jsonResp, err := json.Marshal(response)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		utils.WriteError(w, http.StatusConflict, err.Error())
		return err
	}
	return nil
} // ... other handlers

package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type APIError struct {
	Error string
}

type APIServer struct {
	listenAddr string
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/add_expense", makeHTTPHandleFunc(s.handleAddExpense)).Methods("POST")
	router.HandleFunc("/compare", makeHTTPHandleFunc(s.handleCompareForecastToExpenditure)).Methods("GET")
	log.Printf("Server running on port: %s", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, router); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

func (s *APIServer) handleAddExpense(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleMonthlyForecast(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *APIServer) handleGenerateSummary(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	return nil
}

func (s *APIServer) handleCompareForecastToExpenditure(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "GET" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}
	return WriteJSON(w, http.StatusOK, "Go go gadget!")
}

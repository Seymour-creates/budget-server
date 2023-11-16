package utils

import (
	"encoding/json"
	"errors"
	"github.com/Seymour-creates/budget-server/internal/types"
	"log"
	"net/http"
)

func WriteError(w http.ResponseWriter, httpErr *types.HTTPError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.StatusCode)
	if err := json.NewEncoder(w).Encode(httpErr); err != nil {
		log.Printf("error writing error response: %v", err)
	}
}

func WriteJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // You can make this dynamic if needed
	return json.NewEncoder(w).Encode(data)
}

func ErrorHandler(f types.APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			var httpErr *types.HTTPError
			if errors.As(err, &httpErr) {
				WriteError(w, httpErr)
			}
			log.Printf("########ERROR: %v", err)
			WriteError(w, NewHTTPError(http.StatusInternalServerError, err.Error()))
		}
	}
}

func NewHTTPError(statusCode int, message string) *types.HTTPError {
	return &types.HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

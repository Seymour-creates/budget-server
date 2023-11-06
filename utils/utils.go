package utils

import (
	"encoding/json"
	"github.com/Seymour-creates/budget-server/types"
	"log"
	"net/http"
	"os"
	"strings"
)

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(types.ErrorResponse{Status: status, Message: message}); err != nil {
		log.Printf("error writing error response: %v", err)
	}
}

func ErrorHandler(f types.APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			log.Printf("error: %v", err)
			WriteError(w, http.StatusInternalServerError, err.Error())
		}
	}
}

// LoadConfig returns a map of all environment variables.
func LoadConfig() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}
	return envMap
}

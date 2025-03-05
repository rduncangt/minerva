package api

import (
	"encoding/json"
	"net/http"
)

// jsonResponse writes JSON response with appropriate headers.
func JsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// jsonErrorResponse writes a JSON error response.
func JsonErrorResponse(w http.ResponseWriter, status int, message string) {
	JsonResponse(w, status, map[string]string{"error": message})
}

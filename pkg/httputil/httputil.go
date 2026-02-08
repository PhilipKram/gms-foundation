package httputil

import (
	"encoding/json"
	"net/http"
)

// PaginatedResponse is a standard envelope for paginated list endpoints.
type PaginatedResponse struct {
	Items  interface{} `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

// WriteJSON serializes v as JSON and writes it to w with the given HTTP status code.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError writes a JSON error response of the form {"error": msg}.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

// WritePaginated writes a PaginatedResponse as JSON with the given status code.
func WritePaginated(w http.ResponseWriter, status int, items interface{}, total, limit, offset int) {
	WriteJSON(w, status, PaginatedResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

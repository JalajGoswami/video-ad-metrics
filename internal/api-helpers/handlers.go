package apihelpers

import (
	"encoding/json"
	"net/http"
)

func SuccessResponse(w http.ResponseWriter, status int, result any, message string) {
	if message == "" {
		message = "Request successful"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		map[string]any{"success": true, "message": message, "result": result},
	)
}

func ErrorResponse(w http.ResponseWriter, status int, message string) {
	if message == "" {
		message = "Request failed unexpectedly"
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		map[string]any{"success": false, "message": message},
	)
}

type PaginationOptions struct {
	Limit  int
	Offset int
}

func (p *PaginationOptions) Default() {
	if p.Limit == 0 {
		p.Limit = 25
	}
	if p.Offset == 0 {
		p.Offset = 0
	}
}

type SortOrderOptions struct {
	Order string
}

func (s *SortOrderOptions) Default() {
	if s.Order == "" {
		s.Order = "desc"
	}
}

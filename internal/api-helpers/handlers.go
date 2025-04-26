package apihelpers

import (
	"cmp"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func SuccessResponse(r *http.Request, w http.ResponseWriter, status int, result any, message string) {
	if message == "" {
		message = "Request successful"
	}
	traceID := GetTraceId(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		map[string]any{"success": true, "trace_id": traceID, "message": message, "result": result},
	)
}

func ErrorResponse(r *http.Request, w http.ResponseWriter, status int, message string) {
	if message == "" {
		message = "Request failed unexpectedly"
	}
	traceID := GetTraceId(r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		map[string]any{"success": false, "trace_id": traceID, "message": message},
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

func Pagination(r *http.Request) (PaginationOptions, func(count int, total int) map[string]int, error) {
	opts := PaginationOptions{}
	query := r.URL.Query()
	page, err := strconv.Atoi(cmp.Or(query.Get("page"), "1"))
	if err != nil || page < 1 {
		return opts, nil, errors.New("invalid value for query param `page` provided")
	}
	rows, err := strconv.Atoi(cmp.Or(query.Get("rows"), "25"))
	if err != nil {
		return opts, nil, errors.New("invalid value for query param `rows` provided")
	}
	opts.Limit = rows
	opts.Offset = (page - 1) * rows

	getPaginationObject := func(count int, total int) map[string]int {
		totalPages := total / rows
		if total%rows > 0 {
			totalPages++
		}
		return map[string]int{"page_number": page, "total_pages": totalPages, "total_rows": count, "page_size": rows}
	}
	return opts, getPaginationObject, nil
}

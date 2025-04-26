package apihelpers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type traceIDKey string

const TraceIDKey traceIDKey = "x-trace-id"

// TraceMiddleware adds a trace ID to each request
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a trace ID for this request
		traceID := uuid.New().String()

		// Add trace ID to the request context
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// GetTraceId retrieves the trace ID from the request context
func GetTraceId(r *http.Request) string {
	if traceID, ok := r.Context().Value(TraceIDKey).(string); ok {
		return traceID
	}
	return "-"
}

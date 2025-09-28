package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := uuid.New().String()

			// Wrap ResponseWriter to capture status
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			// Call next
			next.ServeHTTP(ww, r)

			// Log after request is done
			duration := time.Since(start)
			logger.Info("request completed",
				"request_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", strconv.FormatInt(duration.Milliseconds(), 10),
			)

		})
	}
}

// responseWriter wraps http.ResponseWriter to track status codes
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

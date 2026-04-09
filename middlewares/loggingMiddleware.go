package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Capturing the requestId from the context safely
		var requestId string = "unknown"
		if id, ok := r.Context().Value("requestID").(string); ok {
			requestId = id
		}

		// Wrap the ResponseWriter
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		slog.Info("Incoming Request",
			slog.String("Method", r.Method),
			slog.String("Path", r.URL.Path),
			slog.String("RequestID", requestId),
		)

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		slog.Info("Completed Request",
			slog.Int("StatusCode", wrapped.statusCode),
			slog.Duration("Duration", duration),
			slog.String("RequestID", requestId),
		)
	})
}

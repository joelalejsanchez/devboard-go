// Package middlewares provides HTTP middleware helpers.
package middlewares

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader records the status code and writes it to the wrapped response.
func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
		rr.ResponseWriter.WriteHeader(code)
	}
}

// Write writes response data and records the number of bytes written.
func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.WriteHeader(http.StatusOK)
		rr.written = true
	}
	return rr.ResponseWriter.Write(b)
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		// written:        false,
	}
}

// Logger logs request method, path, status code, and response size.
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := newResponseRecorder(w)
			next.ServeHTTP(rec, r)

			// log structurado
			logger.InfoContext(r.Context(), "request completado",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.statusCode),
				slog.Duration("duration", time.Duration(time.Since(start).Milliseconds())),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

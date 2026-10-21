package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err != nil {
					stack := debug.Stack()
					logger.ErrorContext(r.Context(), "panic recuperado",
						slog.Any("error", err),
						slog.String("stack", string(stack)),
						slog.String("path", r.URL.Path),
					)
					http.Error(w, "error interno del servidor", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

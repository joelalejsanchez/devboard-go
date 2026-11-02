// Package middlewares provides HTTP middleware helpers.
package middlewares

import (
	"context"
	"devboard/internal/auth"
	"devboard/internal/domain"
	"devboard/internal/handlers/utils"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const (
	claimsKey contextKey = "auth_claims"
)

// RequireAuth middleware applies token verification and injects claims
func RequireAuth(jwtManager *auth.JWTManager, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := extractBearerToken(r)
			if err != nil {
				utils.RespondError(w, r, logger, domain.ErrUnauthorized)
				return
			}

			claims, err := jwtManager.VerifyToken(tokenString)
			if err != nil {
				utils.RespondError(w, r, logger, domain.ErrUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", domain.ErrUnauthorized
	}
	parts := strings.SplitN(authHeader, " ", 2)

	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", domain.ErrUnauthorized
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", domain.ErrUnauthorized
	}

	return token, nil
}

// GetClaims extracts jwt claims
func GetClaims(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return claims
}

// GetUserID helper to extract userID from context
func GetUserID(r *http.Request) string {
	claims := GetClaims(r)
	if claims == nil {
		return ""
	}
	return claims.UserID
}

// RequireRole middleware verifies authorized roles in endpoint
func RequireRole(role string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				logger.ErrorContext(r.Context(),
					"RequireRole usado sin RequireAuth",
					slog.String("path", r.URL.Path),
				)

				utils.RespondError(w, r, logger, domain.ErrUnauthorized)
				return
			}

			if claims.Role != role {
				utils.RespondError(w, r, logger, domain.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

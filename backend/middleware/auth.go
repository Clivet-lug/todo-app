package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Clivet-lug/todo-app/backend/auth"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

// contextKey is an unexported type so we don't clash with other packages
type contextKey string

const claimsKey contextKey = "claims"

// GetClaims pulls the JWT claims injected by RequireAuth from the request context.
// Returns nil if the middleware wasn't applied (shouldn't happen on protected routes).
func GetClaims(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return claims
}

// RequireAuth validates the Bearer token in the Authorization header.
// On success it injects the claims into the request context and calls next.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header missing",
			})
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header must be: Bearer <token>",
			})
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token",
			})
			return
		}

		// Inject claims into context so handlers can read user identity
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// RequireAdmin builds on RequireAuth and additionally rejects non-admin roles.
func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil || claims.Role != "admin" {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Admin access required",
			})
			return
		}
		next(w, r)
	})
}
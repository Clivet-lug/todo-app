package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Clivet-lug/todo-app/backend/auth"
	"github.com/Clivet-lug/todo-app/backend/models"
	"github.com/Clivet-lug/todo-app/backend/utils"
)

type contextKey string

const claimsKey contextKey = "claims"

func GetClaims(r *http.Request) *auth.Claims {
	claims, _ := r.Context().Value(claimsKey).(*auth.Claims)
	return claims
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Authorization header missing"})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Authorization header must be: Bearer <token>"})
			return
		}
		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			utils.SendJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Invalid or expired token"})
			return
		}
		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if claims := GetClaims(r); claims == nil || claims.Role != "admin" {
			utils.SendJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Message: "Admin access required"})
			return
		}
		next(w, r)
	})
}
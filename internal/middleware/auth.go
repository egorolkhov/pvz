package middleware

import (
	"avito/pkg/auth"
	"context"
	"net/http"
)

func AuthMiddleware(tokenManager auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Missing Authorization header", http.StatusBadRequest)
				return
			}

			role, userID, err := tokenManager.Parse(token)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), "role", role)
			ctx = context.WithValue(ctx, "userID", userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

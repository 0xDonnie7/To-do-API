package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const userIDKey contextKey = "user_id"

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(app.cfg.jwt.secret), nil
		})

		if err != nil || !token.Valid {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), userIDKey, userID))

		next.ServeHTTP(w, r)
	})
}

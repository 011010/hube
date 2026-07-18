// Package middleware provides HTTP middleware for the hube API.
package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

const bearerPrefix = "Bearer "

// Auth returns middleware that requires an "Authorization: Bearer <token>"
// header on every request. Requests to exactly /health and CORS preflight
// (OPTIONS) requests pass through without authentication — preflight requests
// never carry an Authorization header. The token comparison runs in constant
// time to avoid leaking the token through timing attacks.
func Auth(token string) func(http.Handler) http.Handler {
	expected := []byte(token)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			header := r.Header.Get("Authorization")
			got := strings.TrimPrefix(header, bearerPrefix)
			if !strings.HasPrefix(header, bearerPrefix) || subtle.ConstantTimeCompare([]byte(got), expected) != 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

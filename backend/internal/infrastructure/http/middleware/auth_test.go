package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// okHandler reports that the request reached the protected handler.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAuth(t *testing.T) {
	const token = "secret-token"
	handler := Auth(token)(okHandler())

	tests := []struct {
		name       string
		method     string
		path       string
		authHeader string
		wantStatus int
	}{
		{"no authorization header", http.MethodGet, "/api/v1/tasks", "", http.StatusUnauthorized},
		{"wrong token", http.MethodGet, "/api/v1/tasks", "Bearer wrong-token", http.StatusUnauthorized},
		{"missing bearer scheme", http.MethodGet, "/api/v1/tasks", token, http.StatusUnauthorized},
		{"correct token", http.MethodGet, "/api/v1/tasks", "Bearer " + token, http.StatusOK},
		{"health endpoint is open", http.MethodGet, "/health", "", http.StatusOK},
		{"options preflight is open", http.MethodOptions, "/api/v1/tasks", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestAuth_UnauthorizedBody(t *testing.T) {
	handler := Auth("secret-token")(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body map[string]string
	assert.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	assert.Equal(t, "unauthorized", body["error"])
}

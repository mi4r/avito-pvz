package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mi4r/avito-pvz/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	// Helper function to create a valid JWT token
	createToken := func(role string) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": role,
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("secret"))
		return tokenString
	}

	// Helper function to create a request with a token
	createRequest := func(token string) *http.Request {
		req := httptest.NewRequest("GET", "/", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return req
	}

	// Test handler that checks for role in context
	roleCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role")
		if role == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := createRequest("")
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authorization header required")
	})

	t.Run("invalid token format", func(t *testing.T) {
		req := createRequest("invalid.token.format")
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
	})

	t.Run("invalid token signature", func(t *testing.T) {
		// Create token with different secret
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": "admin",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("wrong_secret"))

		req := createRequest(tokenString)
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
	})

	t.Run("valid token with role", func(t *testing.T) {
		tokenString := createToken("admin")
		req := createRequest(tokenString)
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("expired token", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": "admin",
			"exp":  time.Now().Add(-time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("secret"))

		req := createRequest(tokenString)
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
	})

	t.Run("token without role claim", func(t *testing.T) {
		// Create token without role claim
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("secret"))

		req := createRequest(tokenString)
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Empty string is the default value for missing role
	})

	t.Run("unexpected signing method", func(t *testing.T) {
		// Create token with different signing method
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"role": "admin",
			"exp":  time.Now().Add(time.Hour).Unix(),
		})
		tokenString, _ := token.SignedString([]byte("secret"))

		req := createRequest(tokenString)
		w := httptest.NewRecorder()

		middleware.Auth(roleCheckHandler).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

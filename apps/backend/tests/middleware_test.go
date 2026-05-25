package tests

import (
	"incident-flow/backend/internal/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/**
 * Tests for AuthMiddleware and protected routes.
 * These tests cover:
 * - Accessing protected routes with a valid token
 * - Accessing protected routes without a token
 * - Accessing protected routes with an invalid token
 * - Accessing protected routes with a token that has the wrong "Bearer" prefix
 * - Accessing protected routes with an expired token
 * - Accessing protected routes with empty token after Bearer
 * - Accessing protected routes with extra spaces in Authorization header
 */

func TestProtectedRouteWithValidToken(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	// Generate valid token for userID=42
	token, _ := auth.GenerateAuthToken(3)

	req := httptest.NewRequest(
		"GET",
		"/api/profile",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "userID")
	assert.Contains(t, w.Body.String(), "3")
}

func TestProtectedRouteWithoutToken(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestProtectedRouteWithInvalidToken(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestProtectedRouteWithWrongBearerFormat(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "InvalidPrefix some-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestProtectedRouteWithExpiredToken(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	// Generate token that expired 1 hour ago
	expiredToken, _ := auth.GenerateAuthTokenWithSecretAndExpiry(1, []byte("test-secret"), time.Now().Add(-1*time.Hour))

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestProtectedRouteWithEmptyTokenAfterBearer(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestProtectedRouteAuthHeaderWithExtraSpaces(t *testing.T) {
	userRepo := setupTestDB(t)
	router := setupTestRouter(userRepo)

	token, _ := auth.GenerateAuthToken(1)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	// Header with extra spaces: "Bearer  <token>"
	req.Header.Set("Authorization", "Bearer  "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// strings.Fields() handles multiple spaces, so this should still work
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "userID")
}

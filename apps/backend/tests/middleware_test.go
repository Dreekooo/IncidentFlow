package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"incident-flow/backend/internal/auth"
	"incident-flow/backend/internal/repository"
	"net/http"
	"net/http/httptest"
	"os"
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
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Generate valid token for userID=3
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
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestProtectedRouteWithInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestProtectedRouteWithWrongBearerFormat(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "InvalidPrefix some-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestProtectedRouteWithExpiredToken(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Generate token that expired 1 hour ago
	secret := []byte(os.Getenv("JWT_SECRET"))
	expiredToken, _ := auth.GenerateAuthTokenWithSecretAndExpiry(3, secret, time.Now().Add(-1*time.Hour))

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestProtectedRouteWithEmptyTokenAfterBearer(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization header format")
}

func TestProtectedRouteAuthHeaderWithExtraSpaces(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	token, _ := auth.GenerateAuthToken(3)

	req := httptest.NewRequest("GET", "/api/profile", nil)
	// Header with extra spaces: "Bearer  <token>"
	req.Header.Set("Authorization", "Bearer  "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// strings.Fields() handles multiple spaces, so this should still work
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "userID")
}

func TestRequireProjectRole_ForbiddenForMember(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)
	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Create project
	tokenOwner := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Role Test Project", "description": "Testing member restrictions"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)

	var createResp struct {
		Project struct {
			ID       int    `json:"id"`
			JoinCode string `json:"join_code"`
		} `json:"project"`
	}
	err := json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	assert.NoError(t, err)

	// Add a member to the project using the join code
	tokenMember := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	urlJoin := "/api/projects/join"
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)
	assert.Equal(t, http.StatusCreated, wJoin.Code)

	var joinResp struct {
		Request struct {
			ID int `json:"id"`
		} `json:"request"`
	}
	err = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)
	assert.NoError(t, err)

	// Accept the join request as the project owner
	handleBody, _ := json.Marshal(map[string]interface{}{
		"request_id": joinResp.Request.ID,
		"action":     "approve",
	})
	urlHandle := fmt.Sprintf("/api/projects/%d/requests/handle", createResp.Project.ID)
	reqHandle, _ := http.NewRequest(http.MethodPost, urlHandle, bytes.NewBuffer(handleBody))
	reqHandle.Header.Set("Content-Type", "application/json")
	reqHandle.Header.Set("Authorization", "Bearer "+tokenOwner)

	wHandle := httptest.NewRecorder()
	router.ServeHTTP(wHandle, reqHandle)
	assert.Equal(t, http.StatusOK, wHandle.Code)

	// Try to access a route that requires "admin" role with the member's token
	urlRequests := fmt.Sprintf("/api/projects/%d/requests", createResp.Project.ID)
	reqRequests, _ := http.NewRequest(http.MethodGet, urlRequests, nil)
	reqRequests.Header.Set("Authorization", "Bearer "+tokenMember)

	wRequests := httptest.NewRecorder()
	router.ServeHTTP(wRequests, reqRequests)

	assert.Equal(t, http.StatusForbidden, wRequests.Code)
	assert.Contains(t, wRequests.Body.String(), "insufficient project permissions")
}

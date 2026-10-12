package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"incident-flow/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func registerAndLoginTestUser(t *testing.T, router http.Handler) string {
	t.Helper()

	// Register a test user
	email := fmt.Sprintf("test%d@test.com", time.Now().UnixNano())

	registerBody := []byte(`{
	"email": "` + email + `",
	"password": "password123",
	"first_name": "John",
	"last_name": "Doe"
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(registerBody))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Failed to register test user: %s", recorder.Body.String())
	}

	// Log in the test user
	loginBody := []byte(`{
	"email": "` + email + `",
	"password": "password123"
	}`)

	req, _ = http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Failed to log in test user: %s", recorder.Body.String())
	}

	var response struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.NotEmpty(t, response.Token, "Token should not be empty")

	return response.Token
}

func TestCreateProject(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	router := setupTestRouter(userRepo, projectRepo)

	token := registerAndLoginTestUser(t, router)

	projectBody := []byte(`{
		"name": "Test Project",
		"description": "This is a test project"
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("Failed to create project: %s", recorder.Body.String())
	}

	assert.Equal(t, http.StatusCreated, recorder.Code, "Expected status code 201 Created")
}

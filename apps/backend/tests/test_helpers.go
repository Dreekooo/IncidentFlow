package tests

import (
	"bytes"
	"encoding/json"
	"incident-flow/backend/internal/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestLogging captures application and Gin logs for a single test.
// Logs are only printed if the test fails.
func setupTestLogging(t testing.TB) {
	t.Helper()

	var buf bytes.Buffer

	gin.DefaultWriter = &buf
	gin.DefaultErrorWriter = &buf
	logger.SetOutput(&buf)

	t.Cleanup(func() {
		gin.DefaultWriter = os.Stdout
		gin.DefaultErrorWriter = os.Stderr
		logger.SetOutput(os.Stdout)

		if t.Failed() {
			t.Logf("\n--- logs for failed test ---\n%s", buf.String())
		}
	})
}

// registerAndLoginTestUser registers a new test user and logs them in, returning the auth token.
func registerAndLoginTestUser(t *testing.T, router http.Handler, email string, password string) string {
	t.Helper()

	registerBody := []byte(`{
	"email": "` + email + `",
	"password": "` + password + `",
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
	"password": "` + password + `"
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

// getProjectIDFromResponse extracts the project ID from a JSON response body.
func getProjectIDFromResponse(t *testing.T, b []byte) any {
	t.Helper()
	var response struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
	}
	err := json.Unmarshal(b, &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Project.ID == 0 {
		t.Fatalf("Project ID is missing in the response")
	}

	return response.Project.ID
}

// Create a test project and return its ID.
func createTestProject(t *testing.T, router *gin.Engine, token, name, description string) int {
	t.Helper()

	body, _ := json.Marshal(map[string]string{
		"name":        name,
		"description": description,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201 when creating project, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal create project response: %v", err)
	}

	if projectMap, ok := response["project"].(map[string]any); ok {
		if rawID, ok := projectMap["id"]; ok {
			return int(rawID.(float64))
		}
	}

	t.Fatalf("could not extract project ID from response body: %s", recorder.Body.String())
	return 0
}

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"incident-flow/backend/internal/logger"
	"incident-flow/backend/internal/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var jwtSecret = []byte("supersecretkey")

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

// registerAndLoginTestUser registers a new test user and logs them in, returning the auth token and user ID.
func registerAndLoginTestUser(t *testing.T, router http.Handler, email string, password string) (string, int) {
	t.Helper()

	// Register the user
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

	// Log in the user
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

	var response map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	token, ok := response["token"].(string)
	assert.True(t, ok && token != "", "Token should not be empty")

	// Extract user ID from the response
	rawID, ok := response["user_id"].(float64)
	if !ok {
		t.Fatalf("user_id not found in login response: %s", recorder.Body.String())
	}

	return token, int(rawID)
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

// createTestIncident creates a test incident for a given project and returns its ID.
func createTestIncident(t *testing.T, router *gin.Engine, token string, projectID int, title, description string) int {
	t.Helper()

	body, _ := json.Marshal(map[string]any{
		"title":       title,
		"description": description,
		"priority":    models.IncidentPriorityMedium,
		"status":      models.IncidentStatusOpen,
	})

	req, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/projects/%d/incidents", projectID),
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 when creating test incident, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal create incident response: %v", err)
	}

	if rawID, ok := response["id"].(float64); ok {
		return int(rawID)
	}

	t.Fatalf("could not extract incident ID from response body: %s", w.Body.String())
	return 0
}

// joinTestProject allows a member user to join a project using the join code and approves the join request.
func joinTestProject(t *testing.T, router *gin.Engine, ownerToken, memberToken string, projectID int) {
	t.Helper()

	// 1. Owner retrieves the join code for the project
	reqCode, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/join-code", projectID), nil)
	reqCode.Header.Set("Authorization", "Bearer "+ownerToken)
	wCode := httptest.NewRecorder()
	router.ServeHTTP(wCode, reqCode)

	var respCode map[string]any
	_ = json.Unmarshal(wCode.Body.Bytes(), &respCode)
	joinCode, ok := respCode["join_code"].(string)
	if !ok {
		t.Fatalf("failed to get join code from response: %s", wCode.Body.String())
	}

	// 2. Member submits a join request using the join code
	joinBody, _ := json.Marshal(map[string]string{"join_code": joinCode})
	reqJoin, _ := http.NewRequest(http.MethodPost, "/api/projects/join", bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+memberToken)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	if wJoin.Code != http.StatusOK && wJoin.Code != http.StatusCreated {
		t.Fatalf("expected status 200 or 201 when joining project, got %d: %s", wJoin.Code, wJoin.Body.String())
	}

	// 3. Extract request_id from response
	var joinResp map[string]any
	_ = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)

	reqMap, ok := joinResp["request"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'request' object in join response, got: %s", wJoin.Body.String())
	}

	requestID, ok := reqMap["id"].(float64)
	if !ok {
		t.Fatalf("failed to extract request id from join response: %s", wJoin.Body.String())
	}

	// 4. Owner approves the join request
	approveBody, _ := json.Marshal(map[string]any{
		"request_id": int(requestID),
		"action":     "approve",
	})

	reqApprove, _ := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/projects/%d/requests/handle", projectID),
		bytes.NewBuffer(approveBody),
	)
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("Authorization", "Bearer "+ownerToken)

	wApprove := httptest.NewRecorder()
	router.ServeHTTP(wApprove, reqApprove)

	if wApprove.Code != http.StatusOK {
		t.Fatalf("expected status 200 when approving join request, got %d: %s", wApprove.Code, wApprove.Body.String())
	}
}

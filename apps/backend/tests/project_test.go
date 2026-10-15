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

func TestCreateProjectWithoutAuth(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	router := setupTestRouter(userRepo, projectRepo)

	projectBody := []byte(`{
		"name": "Test Project",
		"description": "This is a test project"
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Expected status code 401 Unauthorized")
}

func TestCreateProject_ValidationError(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	router := setupTestRouter(userRepo, projectRepo)

	token := registerAndLoginTestUser(t, router)

	projectBody := []byte(`{
		"description": "This is a test project without a name"
	}`)

	req, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected status code 400 Bad Request")
	var response map[string]interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.Contains(t, response, "error", "Response should contain an error message")
}

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

func TestGetProjectByID_Success(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	router := setupTestRouter(userRepo, projectRepo)

	token := registerAndLoginTestUser(t, router)

	// Create a test project
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

	// Get the project by ID
	projectID := getProjectIDFromResponse(t, recorder.Body.Bytes())

	url := fmt.Sprintf("/api/projects/%d", projectID)

	req, _ = http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestGetProjectByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	router := setupTestRouter(userRepo, projectRepo)

	token := registerAndLoginTestUser(t, router)

	// Attempt to get a non-existent project by ID
	nonExistentProjectID := 676767
	url := fmt.Sprintf("/api/projects/%d", nonExistentProjectID)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code, "Expected status code 404 Not Found")
}

func TestJoinProject_Success(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)

	router := setupTestRouter(userRepo, projectRepo)

	tokenOwner := registerAndLoginTestUser(t, router)

	// Create a test project
	projectBody := []byte(`{
		"name": "Test Project",
		"description": "This is a test project"
	}`)

	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	assert.Equal(t, http.StatusCreated, wCreate.Code, "Expected status code 201 Created")

	var createResp struct {
		Project struct {
			ID       int    `json:"id"`
			JoinCode string `json:"join_code"`
		} `json:"project"`
	}
	err := json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.NotEmpty(t, createResp.Project.JoinCode, "Response should contain project data")

	tokenMember := registerAndLoginTestUser(t, router)

	// Join the project using the join code
	joinBody, _ := json.Marshal(map[string]string{
		"join_code": createResp.Project.JoinCode,
	})

	urlJoin := fmt.Sprintf("/api/projects/%d/join", createResp.Project.ID)
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	assert.Equal(t, http.StatusCreated, wJoin.Code, "Expected status code 201 Created")

	var joinResp struct {
		Message string `json:"message"`
		Request struct {
			ID        int    `json:"id"`
			ProjectID int    `json:"project_id"`
			Status    string `json:"status"`
		} `json:"request"`
	}

	err = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)
	assert.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, "Join request created successfully", joinResp.Message, "Expected success message")
	assert.Equal(t, createResp.Project.ID, joinResp.Request.ProjectID, "Project ID should match")
	assert.Equal(t, "pending", joinResp.Request.Status, "Request status should be pending")
}

func TestReviewJoinRequest_Approve(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	router := setupTestRouter(userRepo, projectRepo)

	tokenOwner := registerAndLoginTestUser(t, router)
	projectBody := []byte(`{"name": "Approve Flow Project", "description": "Test approve"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)
	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	var createResp struct {
		Project struct {
			ID       int    `json:"id"`
			JoinCode string `json:"join_code"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	tokenMember := registerAndLoginTestUser(t, router)
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	urlJoin := fmt.Sprintf("/api/projects/%d/join", createResp.Project.ID)
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)
	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	var joinResp struct {
		Request struct {
			ID int `json:"id"`
		} `json:"request"`
	}
	err := json.Unmarshal(wJoin.Body.Bytes(), &joinResp)
	assert.NoError(t, err)
	assert.NotZero(t, joinResp.Request.ID, "Request ID must not be zero")

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

	if wHandle.Code != http.StatusOK {
		t.Fatalf("HandleJoinRequest failed with status %d: %s", wHandle.Code, wHandle.Body.String())
	}

	assert.Equal(t, http.StatusOK, wHandle.Code)
}

func TestReviewJoinRequest_Reject(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	router := setupTestRouter(userRepo, projectRepo)

	// Owner registers and creates a project
	tokenOwner := registerAndLoginTestUser(t, router)
	projectBody := []byte(`{"name": "Reject Flow Project", "description": "Test reject"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	var createResp struct {
		Project struct {
			ID       int    `json:"id"`
			JoinCode string `json:"join_code"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	// Member registers and sends a join request
	tokenMember := registerAndLoginTestUser(t, router)
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	urlJoin := fmt.Sprintf("/api/projects/%d/join", createResp.Project.ID)
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	var joinResp struct {
		Request struct {
			ID int `json:"id"`
		} `json:"request"`
	}
	err := json.Unmarshal(wJoin.Body.Bytes(), &joinResp)
	assert.NoError(t, err)
	assert.NotZero(t, joinResp.Request.ID, "Request ID must not be zero")

	// Owner rejects the join request
	handleBody, _ := json.Marshal(map[string]interface{}{
		"request_id": joinResp.Request.ID,
		"action":     "reject",
	})

	urlHandle := fmt.Sprintf("/api/projects/%d/requests/handle", createResp.Project.ID)
	reqHandle, _ := http.NewRequest(http.MethodPost, urlHandle, bytes.NewBuffer(handleBody))
	reqHandle.Header.Set("Content-Type", "application/json")
	reqHandle.Header.Set("Authorization", "Bearer "+tokenOwner)

	wHandle := httptest.NewRecorder()
	router.ServeHTTP(wHandle, reqHandle)

	if wHandle.Code != http.StatusOK {
		t.Fatalf("HandleJoinRequest failed with status %d: %s", wHandle.Code, wHandle.Body.String())
	}

	assert.Equal(t, http.StatusOK, wHandle.Code)
}

func TestJoinProject_AlreadyMember(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	router := setupTestRouter(userRepo, projectRepo)

	tokenOwner := registerAndLoginTestUser(t, router)
	projectBody := []byte(`{"name": "Already Member Project", "description": "Test already member"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	var createResp struct {
		Project struct {
			ID       int    `json:"id"`
			JoinCode string `json:"join_code"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})

	urlJoin := fmt.Sprintf("/api/projects/%d/join", createResp.Project.ID)
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenOwner)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	assert.Equal(t, http.StatusConflict, wJoin.Code, "Expected status code 409 Conflict")
}

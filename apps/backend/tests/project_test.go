package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"incident-flow/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestCreateProject(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	token, _ := registerAndLoginTestUser(t, router, "test@test.com", "password123")

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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	token, _ := registerAndLoginTestUser(t, router, "test@test.com", "password123")

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

func TestGetProjectByID_Success(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	token, _ := registerAndLoginTestUser(t, router, "test@test.com", "password123")

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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	token, _ := registerAndLoginTestUser(t, router, "test@test.com", "password123")

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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")

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

	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")

	// Join the project using the join code
	joinBody, _ := json.Marshal(map[string]string{
		"join_code": createResp.Project.JoinCode,
	})

	urlJoin := "/api/projects/join"
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

func TestJoinProject_InvalidJoinCode(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")

	// Try to join a project with an invalid join code
	joinBody, _ := json.Marshal(map[string]string{
		"join_code": "invalid-join-code",
	})

	urlJoin := "/api/projects/join"
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	assert.Equal(t, http.StatusNotFound, wJoin.Code)
}

func TestReviewJoinRequest_Approve(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
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

	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	urlJoin := "/api/projects/join"
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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner registers and creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
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
	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	urlJoin := "/api/projects/join"
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
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
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

	urlJoin := "/api/projects/join"
	reqJoin, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenOwner)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	assert.Equal(t, http.StatusConflict, wJoin.Code, "Expected status code 409 Conflict")
}

func TestJoinProject_AlreadyRequested(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")

	projectBody := []byte(`{"name": "Already Requested Project", "description": "Test already requested"}`)
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

	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})

	urlJoin := "/api/projects/join"

	// First join request
	reqJoin1, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin1.Header.Set("Content-Type", "application/json")
	reqJoin1.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin1 := httptest.NewRecorder()
	router.ServeHTTP(wJoin1, reqJoin1)

	assert.Equal(t, http.StatusCreated, wJoin1.Code, "Expected status code 201 Created for first join request")

	// Attempt to send a second join request with the same join code
	reqJoin2, _ := http.NewRequest(http.MethodPost, urlJoin, bytes.NewBuffer(joinBody))
	reqJoin2.Header.Set("Content-Type", "application/json")
	reqJoin2.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin2 := httptest.NewRecorder()
	router.ServeHTTP(wJoin2, reqJoin2)

	assert.Equal(t, http.StatusConflict, wJoin2.Code, "Expected status code 409 Conflict for duplicate join request")
}

func TestRemoveProjectMember_Success(t *testing.T) {
	db := setupTestDB(t)

	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner registers and creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	projectBody := []byte(`{"name": "Remove Member Project", "description": "Test remove member"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(projectBody))
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

	// Member registers and sends a join request
	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
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
			ID     int `json:"id"`
			UserID int `json:"user_id"`
		} `json:"request"`
	}
	err = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)
	assert.NoError(t, err)

	// Owner approves the join request
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

	// Owner removes the member from the project
	urlRemove := fmt.Sprintf("/api/projects/%d/members/%d", createResp.Project.ID, joinResp.Request.UserID)
	reqRemove, _ := http.NewRequest(http.MethodDelete, urlRemove, nil)
	reqRemove.Header.Set("Authorization", "Bearer "+tokenOwner)

	wRemove := httptest.NewRecorder()
	router.ServeHTTP(wRemove, reqRemove)
	assert.Equal(t, http.StatusOK, wRemove.Code)

	// Verify that the deleted member can no longer access the project
	urlGetProject := fmt.Sprintf("/api/projects/%d", createResp.Project.ID)
	reqGetProject, _ := http.NewRequest(http.MethodGet, urlGetProject, nil)
	reqGetProject.Header.Set("Authorization", "Bearer "+tokenMember)

	wGetProject := httptest.NewRecorder()
	router.ServeHTTP(wGetProject, reqGetProject)

	assert.Equal(t, http.StatusForbidden, wGetProject.Code)
}

func TestGetProjectMembers_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner registers and creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Members List Test", "description": "Testing members endpoint"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)

	var createResp struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	// Owner retrieves the list of project members
	urlMembers := fmt.Sprintf("/api/projects/%d/members", createResp.Project.ID)
	reqMembers, _ := http.NewRequest(http.MethodGet, urlMembers, nil)
	reqMembers.Header.Set("Authorization", "Bearer "+tokenOwner)

	wMembers := httptest.NewRecorder()
	router.ServeHTTP(wMembers, reqMembers)

	assert.Equal(t, http.StatusOK, wMembers.Code)
	assert.Contains(t, wMembers.Body.String(), "owner")
}

func TestGetProjectMembers_ForbiddenForNonMember(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner registers and creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Private Members List", "description": "Security test"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	var createResp struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	// Another user (not a member) tries to access the project members list
	tokenOther, _ := registerAndLoginTestUser(t, router, "other@test.com", "password123!")
	urlMembers := fmt.Sprintf("/api/projects/%d/members", createResp.Project.ID)
	reqMembers, _ := http.NewRequest(http.MethodGet, urlMembers, nil)
	reqMembers.Header.Set("Authorization", "Bearer "+tokenOther)

	wMembers := httptest.NewRecorder()
	router.ServeHTTP(wMembers, reqMembers)

	assert.Equal(t, http.StatusForbidden, wMembers.Code)
}

func TestUpdateProject_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Original Name", "description": "Original Description"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)
	assert.Equal(t, http.StatusCreated, wCreate.Code)

	var createResp struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	// Owner updates the project
	updateBody := []byte(`{"name": "Updated Project Name", "description": "Updated Description"}`)
	urlUpdate := fmt.Sprintf("/api/projects/%d", createResp.Project.ID)
	reqUpdate, _ := http.NewRequest(http.MethodPut, urlUpdate, bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	reqUpdate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)

	assert.Equal(t, http.StatusOK, wUpdate.Code)

	var updateResp struct {
		Project struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"project"`
	}
	err := json.Unmarshal(wUpdate.Body.Bytes(), &updateResp)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Project Name", updateResp.Project.Name)
	assert.Equal(t, "Updated Description", updateResp.Project.Description)
}

func TestUpdateMemberRole_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and a member joins
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Role Change Project", "description": "Testing role update"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
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

	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	reqJoin, _ := http.NewRequest(http.MethodPost, "/api/projects/join", bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	var joinResp struct {
		Request struct {
			ID     int `json:"id"`
			UserID int `json:"user_id"`
		} `json:"request"`
	}
	_ = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)

	approveBody, _ := json.Marshal(map[string]interface{}{"request_id": joinResp.Request.ID, "action": "approve"})
	urlHandle := fmt.Sprintf("/api/projects/%d/requests/handle", createResp.Project.ID)
	reqApprove, _ := http.NewRequest(http.MethodPost, urlHandle, bytes.NewBuffer(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("Authorization", "Bearer "+tokenOwner)

	wApprove := httptest.NewRecorder()
	router.ServeHTTP(wApprove, reqApprove)
	assert.Equal(t, http.StatusOK, wApprove.Code)

	// Owner updates the member's role to 'admin'
	roleBody := []byte(`{"role": "admin"}`)
	urlRole := fmt.Sprintf("/api/projects/%d/members/%d/role", createResp.Project.ID, joinResp.Request.UserID)
	reqRole, _ := http.NewRequest(http.MethodPatch, urlRole, bytes.NewBuffer(roleBody))
	reqRole.Header.Set("Content-Type", "application/json")
	reqRole.Header.Set("Authorization", "Bearer "+tokenOwner)

	wRole := httptest.NewRecorder()
	router.ServeHTTP(wRole, reqRole)

	assert.Equal(t, http.StatusOK, wRole.Code)
}

func TestUpdateMemberRole_ForbiddenForMember(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Forbidden Role Project", "description": "Testing permissions"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
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

	// Member registers and joins the project
	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	reqJoin, _ := http.NewRequest(http.MethodPost, "/api/projects/join", bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	var joinResp struct {
		Request struct {
			ID     int `json:"id"`
			UserID int `json:"user_id"`
		} `json:"request"`
	}
	_ = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)

	approveBody, _ := json.Marshal(map[string]interface{}{"request_id": joinResp.Request.ID, "action": "approve"})
	urlHandle := fmt.Sprintf("/api/projects/%d/requests/handle", createResp.Project.ID)
	reqApprove, _ := http.NewRequest(http.MethodPost, urlHandle, bytes.NewBuffer(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("Authorization", "Bearer "+tokenOwner)
	wApprove := httptest.NewRecorder()
	router.ServeHTTP(wApprove, reqApprove)

	// Member attempts to update their own role to 'admin'
	roleBody := []byte(`{"role": "admin"}`)
	urlRole := fmt.Sprintf("/api/projects/%d/members/%d/role", createResp.Project.ID, joinResp.Request.UserID)
	reqRole, _ := http.NewRequest(http.MethodPatch, urlRole, bytes.NewBuffer(roleBody))
	reqRole.Header.Set("Content-Type", "application/json")
	reqRole.Header.Set("Authorization", "Bearer "+tokenMember)

	wRole := httptest.NewRecorder()
	router.ServeHTTP(wRole, reqRole)

	// The member should receive a 403 Forbidden response since they are not authorized to change roles
	assert.Equal(t, http.StatusForbidden, wRole.Code)
}

func TestLeaveProject_Success(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project
	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Leave Test Project", "description": "Testing leave functionality"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
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

	// Member registers, joins the project, and gets approved
	tokenMember, _ := registerAndLoginTestUser(t, router, "member@test.com", "password123!")
	joinBody, _ := json.Marshal(map[string]string{"join_code": createResp.Project.JoinCode})
	reqJoin, _ := http.NewRequest(http.MethodPost, "/api/projects/join", bytes.NewBuffer(joinBody))
	reqJoin.Header.Set("Content-Type", "application/json")
	reqJoin.Header.Set("Authorization", "Bearer "+tokenMember)

	wJoin := httptest.NewRecorder()
	router.ServeHTTP(wJoin, reqJoin)

	var joinResp struct {
		Request struct {
			ID int `json:"id"`
		} `json:"request"`
	}
	_ = json.Unmarshal(wJoin.Body.Bytes(), &joinResp)

	// Owner approves the join request
	approveBody, _ := json.Marshal(map[string]interface{}{"request_id": joinResp.Request.ID, "action": "approve"})
	urlHandle := fmt.Sprintf("/api/projects/%d/requests/handle", createResp.Project.ID)
	reqApprove, _ := http.NewRequest(http.MethodPost, urlHandle, bytes.NewBuffer(approveBody))
	reqApprove.Header.Set("Content-Type", "application/json")
	reqApprove.Header.Set("Authorization", "Bearer "+tokenOwner)

	wApprove := httptest.NewRecorder()
	router.ServeHTTP(wApprove, reqApprove)
	assert.Equal(t, http.StatusOK, wApprove.Code)

	// Member leaves the project
	urlLeave := fmt.Sprintf("/api/projects/%d/leave", createResp.Project.ID)
	reqLeave, _ := http.NewRequest(http.MethodPost, urlLeave, nil)
	reqLeave.Header.Set("Authorization", "Bearer "+tokenMember)

	wLeave := httptest.NewRecorder()
	router.ServeHTTP(wLeave, reqLeave)

	assert.Equal(t, http.StatusOK, wLeave.Code)

	// Verify that the member can no longer access the project after leaving
	urlGet := fmt.Sprintf("/api/projects/%d", createResp.Project.ID)
	reqGet, _ := http.NewRequest(http.MethodGet, urlGet, nil)
	reqGet.Header.Set("Authorization", "Bearer "+tokenMember)

	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)

	assert.Equal(t, http.StatusForbidden, wGet.Code)
}

func TestLeaveProject_OwnerForbidden(t *testing.T) {
	db := setupTestDB(t)
	userRepo := repository.NewUserRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	tokenOwner, _ := registerAndLoginTestUser(t, router, "owner@test.com", "password123!")
	createBody := []byte(`{"name": "Owner Leave Project", "description": "Testing owner restriction"}`)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/projects", bytes.NewBuffer(createBody))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+tokenOwner)

	wCreate := httptest.NewRecorder()
	router.ServeHTTP(wCreate, reqCreate)

	var createResp struct {
		Project struct {
			ID int `json:"id"`
		} `json:"project"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)

	// Owner attempts to leave the project
	urlLeave := fmt.Sprintf("/api/projects/%d/leave", createResp.Project.ID)
	reqLeave, _ := http.NewRequest(http.MethodPost, urlLeave, nil)
	reqLeave.Header.Set("Authorization", "Bearer "+tokenOwner)

	wLeave := httptest.NewRecorder()
	router.ServeHTTP(wLeave, reqLeave)

	// The owner should receive a 403 Forbidden response since they cannot leave their own project
	assert.Equal(t, http.StatusForbidden, wLeave.Code)
	assert.Contains(t, wLeave.Body.String(), "project owners cannot remove themselves")
}

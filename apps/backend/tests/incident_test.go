package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"incident-flow/backend/internal/models"
	"incident-flow/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

/*
 * Test cases for incident creation
 */

func TestCreateIncident(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Create a new incident
	createIncidentBody := map[string]interface{}{
		"title":       "Critical Server Crash",
		"description": "Database server stopped responding under load",
		"priority":    models.IncidentPriorityLow,
		"status":      models.IncidentStatusOpen,
	}
	jsonBody, _ := json.Marshal(createIncidentBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/incidents", projectID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusCreated, recorder.Code, "Expected status code 201 Created")

	var createdIncident models.Incident
	if err := json.Unmarshal(recorder.Body.Bytes(), &createdIncident); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	assert.Equal(t, createIncidentBody["status"], createdIncident.Status, "Expected status to be set")
}

func TestCreateIncident_Unauthorized(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Attempt to create an incident without authentication
	createIncidentBody := map[string]interface{}{
		"title":       "Unauthorized Incident",
		"description": "This should not be created",
		"priority":    models.IncidentPriorityLow,
		"status":      models.IncidentStatusOpen,
	}
	jsonBody, _ := json.Marshal(createIncidentBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/incidents", 9999), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code, "Expected status code 401 Unauthorized")
}

func TestCreateIncident_ForbiddenForNonMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Create a project with a different user
	otherToken, _ := registerAndLoginTestUser(t, router, "otheruser@example.com", "password123!")

	// Attempt to create an incident in the other user's project
	createIncidentBody := map[string]interface{}{
		"title":       "Forbidden Incident",
		"description": "This should not be created",
		"priority":    models.IncidentPriorityLow,
		"status":      models.IncidentStatusOpen,
	}
	jsonBody, _ := json.Marshal(createIncidentBody)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/incidents", projectID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+otherToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected status code 403 Forbidden")
}

func TestCreateIncident_ValidationError(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Attempt to create an incident with missing required fields
	createIncidentBody := map[string]interface{}{
		"title": "Incomplete Incident",
		// Missing description, priority, and status
	}
	jsonBody, _ := json.Marshal(createIncidentBody)

	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/projects/%d/incidents", projectID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected status code 400 Bad Request")
}

/*
 * Test cases for incident retrieval
 */

func TestGetIncidentByID_Success(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Attempt to retrieve the incident by ID
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestGetIncidentByID_NotFound(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Attempt to retrieve a non-existent incident by ID
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, 9999), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code, "Expected status code 404 Not Found")
}

func TestGetIncidentByID_ForbiddenForNonMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Register and login a non-member user
	nonMemberToken, _ := registerAndLoginTestUser(t, router, "nonmember@example.com", "password123!")

	// Attempt to retrieve the incident by ID as a non-member
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID), nil)
	req.Header.Set("Authorization", "Bearer "+nonMemberToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected status code 403 Forbidden")
}

/*
 * Test cases for incident listing
 */

func TestListIncidents_Success(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	createTestIncident(t, router, ownerToken, projectID, "Test Incident 1", "This is first test incident.")
	createTestIncident(t, router, ownerToken, projectID, "Test Incident 2", "This is second test incident.")

	// Attempt to list incidents for the project
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestListIncidents_Empty(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Attempt to list incidents for the project with no incidents
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestListIncidents_ForbiddenForNonMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Register and login a non-member user
	nonMemberToken, _ := registerAndLoginTestUser(t, router, "nonmember@example.com", "password123!")

	// Attempt to list incidents for the project as a non-member
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/projects/%d/incidents", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+nonMemberToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected status code 403 Forbidden")
}

/*
 * Test cases for incident update
 */

func TestUpdateIncident_Success(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Attempt to update the incident
	updateIncidentBody := map[string]interface{}{
		"title":       "Updated Incident Title",
		"description": "Updated description",
		"priority":    models.IncidentPriorityHigh,
		"status":      models.IncidentStatusInProgress,
	}
	jsonBody, _ := json.Marshal(updateIncidentBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestUpdateIncident_ForbiddenForMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and an incident
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "Password123!")
	projectID := createTestProject(t, router, ownerToken, "Incident Project", "Desc")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Original Incident", "Desc")

	// Member registers and joins the project
	memberToken, _ := registerAndLoginTestUser(t, router, "member@example.com", "Password123!")
	joinTestProject(t, router, ownerToken, memberToken, projectID)

	// Member attempts to update the incident
	updateBody, _ := json.Marshal(map[string]any{
		"title":       "Hacked Title",
		"description": "Unauthorized change",
		"priority":    models.IncidentPriorityHigh,
		"status":      models.IncidentStatusResolved,
	})

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID),
		bytes.NewBuffer(updateBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+memberToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected status code 403 Forbidden")
}

func TestUpdateIncident_NotFound(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Attempt to update a non-existent incident
	updateIncidentBody := map[string]interface{}{
		"title":       "Non-existent Incident",
		"description": "This incident does not exist",
		"priority":    models.IncidentPriorityLow,
		"status":      models.IncidentStatusOpen,
	}

	jsonBody, _ := json.Marshal(updateIncidentBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, 9999), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusNotFound, recorder.Code, "Expected status code 404 Not Found")
}

/*
 * Test cases for incident assignment
 */

func TestAssignIncident_Success(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and an incident
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Member registers and joins the project
	memberToken, memberUserID := registerAndLoginTestUser(t, router, "member@example.com", "password123!")
	joinTestProject(t, router, ownerToken, memberToken, projectID)

	// Owner assigns the incident to the member
	assignIncidentBody := map[string]any{
		"user_id": memberUserID,
	}
	jsonBody, _ := json.Marshal(assignIncidentBody)

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/projects/%d/incidents/%d/assign", projectID, incidentID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestAssignIncident_UserNotMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and an incident
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Non-member registers
	_, nonMemberUserID := registerAndLoginTestUser(t, router, "nonmember@example.com", "password123!")

	// Owner attempts to assign the incident to the non-member
	assignIncidentBody := map[string]any{
		"user_id": nonMemberUserID,
	}
	jsonBody, _ := json.Marshal(assignIncidentBody)

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/projects/%d/incidents/%d/assign", projectID, incidentID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected status code 400 Bad Request for non-member assignment")
}

func TestAssignIncident_Unassign(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and an incident
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Member registers and joins the project
	memberToken, memberUserID := registerAndLoginTestUser(t, router, "member@example.com", "password123!")
	joinTestProject(t, router, ownerToken, memberToken, projectID)

	// Owner assigns the incident to the member
	assignIncidentBody := map[string]any{
		"user_id": memberUserID,
	}
	jsonBody, _ := json.Marshal(assignIncidentBody)

	req, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/projects/%d/incidents/%d/assign", projectID, incidentID),
		bytes.NewBuffer(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")

	// Now unassign the incident by sending null user_id
	unassignIncidentBody := map[string]any{
		"user_id": nil,
	}
	jsonUnassignBody, _ := json.Marshal(unassignIncidentBody)

	reqUnassign, _ := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/api/projects/%d/incidents/%d/assign", projectID, incidentID),
		bytes.NewBuffer(jsonUnassignBody),
	)
	reqUnassign.Header.Set("Content-Type", "application/json")
	reqUnassign.Header.Set("Authorization", "Bearer "+ownerToken)

	router.ServeHTTP(recorder, reqUnassign)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

/*
 * Test cases for incident deletion
 */

func TestDeleteIncident_Success(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Attempt to delete the incident
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID), nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code, "Expected status code 200 OK")
}

func TestDeleteIncident_ForbiddenForMember(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Owner creates a project and an incident
	ownerToken, _ := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")
	incidentID := createTestIncident(t, router, ownerToken, projectID, "Test Incident", "This is a test incident.")

	// Member registers and joins the project
	memberToken, _ := registerAndLoginTestUser(t, router, "member@example.com", "password123!")
	joinTestProject(t, router, ownerToken, memberToken, projectID)

	// Attempt to delete the incident as a member (should be forbidden)
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/projects/%d/incidents/%d", projectID, incidentID), nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusForbidden, recorder.Code, "Expected status code 403 Forbidden")
}

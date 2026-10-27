package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/models"
	
	"github.com/stretchr/testify/assert"
)

func TestCreateIncident(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	// Register and login a test user to get an auth token
	ownerToken := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
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
	ownerToken := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
	projectID := createTestProject(t, router, ownerToken, "Test Project", "This is a test project.")

	// Create a project with a different user
	otherToken := registerAndLoginTestUser(t, router, "otheruser@example.com", "password123!")

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
	ownerToken := registerAndLoginTestUser(t, router, "owner@example.com", "password123!")
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


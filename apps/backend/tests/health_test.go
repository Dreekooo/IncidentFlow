package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"incident-flow/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	database := setupTestDB(t)

	userRepo := repository.NewUserRepository(database)
	projectRepo := repository.NewProjectRepository(database)
	incidentRepo := repository.NewIncidentRepository(database)

	router := setupTestRouter(userRepo, projectRepo, incidentRepo)

	req, _ := http.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

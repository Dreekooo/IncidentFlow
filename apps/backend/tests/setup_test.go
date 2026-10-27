package tests

import (
	"database/sql"
	"testing"

	"incident-flow/backend/internal/db"
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/server"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

/**
 * Test setup utilities for the backend tests.
 * This file includes:
 * - Initialization of the testing environment (loading .env, setting Gin to test mode)
 * - A helper function to set up a test database connection and return a UserRepository
 * - A helper function to set up the Gin router with the necessary handlers for testing
 *
 * These utilities are used across multiple test files to ensure a consistent testing environment.
 */

// init sets up the testing environment by loading environment variables and configuring Gin for test mode.
func init() {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Load("../.env", ".env")
}

// setupTestDB initializes a test database connection, cleans all tables, and returns *sql.DB.
func setupTestDB(t testing.TB) *sql.DB {
	setupTestLogging(t)

	database, err := db.Connect()
	if err != nil {
		t.Fatal("failed to connect to database", err)
	}

	if err := db.InitDB(database); err != nil {
		t.Fatal("failed to initialize database", err)
	}

	_, err = database.Exec("TRUNCATE TABLE users, projects, project_members, incidents RESTART IDENTITY CASCADE;")
	if err != nil {
		t.Fatal("failed to truncate test tables:", err)
	}

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

// setupTestRouter initializes the Gin router with the necessary handlers for testing.
func setupTestRouter(
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
	incidentRepo repository.IncidentRepository,
) *gin.Engine {
	authHandler := handlers.NewAuthHandler(userRepo)
	projectHandler := handlers.NewProjectHandler(projectRepo)
	incidentHandler := handlers.NewIncidentHandler(incidentRepo)

	return server.SetupRouter(authHandler, projectHandler, incidentHandler, projectRepo)
}
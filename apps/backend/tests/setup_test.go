package tests

import (
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

// setupTestDB initializes a test database connection and returns a UserRepository.
func setupTestDB(t testing.TB) *repository.UserRepository {
	database, err := db.Connect()
	if err != nil {
		t.Fatal("failed to connect to database", err)
	}

	database.Exec("TRUNCATE TABLE users;")

	return repository.NewUserRepository(database)
}

// setupTestRouter initializes the Gin router with the necessary handlers for testing.
func setupTestRouter(userRepo *repository.UserRepository) *gin.Engine {
	authHandler := handlers.NewAuthHandler(userRepo)
	return server.SetupRouter(authHandler)
}

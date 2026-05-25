package tests

import (
	"incident-flow/backend/internal/db"
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/server"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Load("./apps/backend/.env")
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

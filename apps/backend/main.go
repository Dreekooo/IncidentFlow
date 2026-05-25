package main

import (
	"log"

	"github.com/joho/godotenv"

	"incident-flow/backend/internal/db"
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/logger"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/server"
)

func main() {
	// Load .env file (optional, useful for local development)
	_ = godotenv.Load()

	// Logger inicjalizuje się automatycznie przez init()
	logger.Info("application starting")

	database, err := db.Connect()
	if err != nil {
		logger.Error("failed to connect to database", "error", err.Error())
		panic(err)
	}

	defer database.Close()

	// Initialize database schema
	if err := db.InitDB(database); err != nil {
		logger.Error("failed to initialize database", "error", err.Error())
		panic(err)
	}

	logger.Info("database initialized successfully")

	// Initialize repositories and handlers
	userRepo := repository.NewUserRepository(database)
	authHandler := handlers.NewAuthHandler(userRepo)

	r := server.SetupRouter(authHandler)

	logger.Info("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Error("failed to start server", "error", err.Error())
		log.Fatalf("failed to start server: %v", err)
	}
}

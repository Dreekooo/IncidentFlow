package main

import (
	"log"

	"incident-flow/backend/internal/db"
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/server"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		panic(err)
	}

	defer database.Close()

	// Initialize database schema
	if err := db.InitDB(database); err != nil {
		panic(err)
	}

	// Initialize repositories and handlers
	userRepo := repository.NewUserRepository(database)
	authHandler := handlers.NewAuthHandler(userRepo)

	r := server.SetupRouter(authHandler)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

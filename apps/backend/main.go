package main

import (
	"log"
	"net/http"

	"incident-flow/backend/internal/db"
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/repository"

	"github.com/gin-gonic/gin"
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

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		err := database.Ping()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error", "message": "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok", "message": "Backend is healthy",
		})
	})

	r.POST("/register", authHandler.Register)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

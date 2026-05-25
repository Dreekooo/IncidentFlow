package server

import (
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(authHandler *handlers.AuthHandler) *gin.Engine {
	r := gin.Default()
	profileHandler := handlers.NewProfileHandler()

	// Apply global middleware
	r.Use(middleware.LoggingMiddleware())

	r.GET("/health", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "Backend is healthy",
		})
	})

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	// Protected routes (require valid JWT)
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", profileHandler.Profile)
	}

	return r
}

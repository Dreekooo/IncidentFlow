package main

import (
	"net/http"

	"incident-flow/backend/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		panic(err)
	}

	defer database.Close()

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

	r.Run(":8080")
}

package handlers

import (
	"incident-flow/backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct{}

func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

func (h *ProfileHandler) Profile(c *gin.Context) {
	userID := middleware.GetUserIDFromContext(c)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile data",
		"userID":  userID,
	})
}

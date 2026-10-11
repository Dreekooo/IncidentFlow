package handlers

import (
	"incident-flow/backend/internal/models"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectRepo repository.ProjectRepository
}

func NewProjectHandler(projectRepo repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo}
}

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID := userIDVal.(int)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var joinCode string

	for {
        code := utils.GenerateRandomJoinCode()
        existing, err := h.projectRepo.GetByJoinCode(c.Request.Context(), code)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
            return
        }
        
        if existing == nil {
            joinCode = code
            break
        }
    }

	project := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		JoinCode:    joinCode,
		OwnerID:     userID,
	}

	if err := h.projectRepo.CreateProject(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": project,
	})
}

func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	var request struct {
		ID int `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), request.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}
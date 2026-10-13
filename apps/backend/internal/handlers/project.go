package handlers

import (
	"incident-flow/backend/internal/models"
	"incident-flow/backend/internal/repository"
	"incident-flow/backend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Project handlers
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

// POST /api/projects
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Project created successfully",
		"project": project,
	})
}

// GET /api/projects/:id
func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	var request struct {
		ID int `uri:"id" binding:"required"`
	}

	if err := c.ShouldBindUri(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), request.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"project": project,
	})
}

// JoinProjectRequest handlers
type JoinProjectRequest struct {
	JoinCode string `json:"join_code" binding:"required"`
}

// POST /api/projects/:id/join
func (h *ProjectHandler) JoinProject(c *gin.Context) {
	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	userID := userIDVal.(int)

	project, err := h.projectRepo.GetByJoinCode(c.Request.Context(), c.Param("join_code"))
	if err != nil || project == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid join code or database error"})
		return
	}

	isMember, err := h.projectRepo.IsProjectMember(c.Request.Context(), project.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if isMember {
		c.JSON(http.StatusConflict, gin.H{"error": "user is already a member of the project"})
		return
	}

	joinRequest := &models.ProjectJoinRequest{
		ProjectID: project.ID,
		UserID:    userID,
	}

	if err := h.projectRepo.CreateJoinRequest(c.Request.Context(), joinRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create join request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Join request created successfully",
	})
}

// GET /api/projects/:id/requests
func (h *ProjectHandler) GetJoinRequests(c *gin.Context) {
	userID := c.GetInt("user_id")
	projectID, _ := strconv.Atoi(c.Param("id"))

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if project.OwnerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only project owner can view join requests"})
		return
	}

	requests, err := h.projectRepo.GetPendingJoinRequests(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve join requests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"join_requests": requests,
	})
}

type HandleJoinRequestInput struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
}

// POST /api/projects/:id/requests/handle
func (h *ProjectHandler) HandleJoinRequest(c *gin.Context) {
	UserID := c.GetInt("user_id")
	requestID, _ := strconv.Atoi(c.Param("id"))

	var input HandleJoinRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	joinRequest, err := h.projectRepo.GetJoinRequestByID(c.Request.Context(), requestID)
	if err != nil || joinRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "join request not found"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), joinRequest.ProjectID)
	if err != nil || project.OwnerID != UserID {
		c.JSON(http.StatusNotFound, gin.H{"error": "unauthorized: only project owner can handle join requests"})
		return
	}

	switch input.Action {
		case "approve":
			err = h.projectRepo.UpdateJoinRequestStatus(c.Request.Context(), requestID, "approved")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve join request"})
				return
			}
			err = h.projectRepo.AddProjectMember(c.Request.Context(), joinRequest.ProjectID, joinRequest.UserID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add project member"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "join request approved"})
		case "reject":
			err = h.projectRepo.UpdateJoinRequestStatus(c.Request.Context(), requestID, "rejected")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject join request"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "join request rejected"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
	}
}

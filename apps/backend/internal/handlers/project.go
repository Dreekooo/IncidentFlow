package handlers

import (
	"database/sql"
	"errors"
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
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	project, err := h.projectRepo.GetByID(c.Request.Context(), id)
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

// GET /api/projects/:id/join-code
func (h *ProjectHandler) GetJoinCode(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"join_code": project.JoinCode,
	})
}

// JoinProjectRequest handlers
type JoinProjectRequest struct {
	JoinCode string `json:"join_code" binding:"required"`
}

// POST /api/projects/join
func (h *ProjectHandler) JoinProject(c *gin.Context) {
	userIDVal, exists := c.Get("userID")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found in context"})
		return
	}

	userID := userIDVal.(int)

	var req struct {
		JoinCode string `json:"join_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	project, err := h.projectRepo.GetByJoinCode(c.Request.Context(), req.JoinCode)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid join code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid join code"})
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

	hasPending, err := h.projectRepo.HasPendingJoinRequest(c.Request.Context(), project.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if hasPending {
		c.JSON(http.StatusConflict, gin.H{"error": "join request already submitted"})
		return
	}

	joinRequest := &models.ProjectJoinRequest{
		ProjectID: project.ID,
		UserID:    userID,
		Status:    "pending",
	}

	if err := h.projectRepo.CreateJoinRequest(c.Request.Context(), joinRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create join request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Join request created successfully",
		"request": joinRequest,
	})
}

// GET /api/projects/:id/requests
func (h *ProjectHandler) GetJoinRequests(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
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
	RequestID int    `json:"request_id" binding:"required"`
	Action    string `json:"action" binding:"required,oneof=approve reject"`
}

// POST /api/projects/:id/requests/handle
func (h *ProjectHandler) HandleJoinRequest(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var input HandleJoinRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	ctx := c.Request.Context()

	project, err := h.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	joinRequest, err := h.projectRepo.GetJoinRequestByID(ctx, input.RequestID)
	if err != nil || joinRequest == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "join request not found"})
		return
	}

	dbStatus := "accepted"
	if input.Action == "reject" {
		dbStatus = "rejected"
	}

	if err := h.projectRepo.UpdateJoinRequestStatus(ctx, input.RequestID, dbStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update request status"})
		return
	}

	if input.Action == "approve" {
		if err := h.projectRepo.AddProjectMember(ctx, joinRequest.ProjectID, joinRequest.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add project member"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "join request processed successfully"})
}

// GET /api/projects/:id/members
func (h *ProjectHandler) GetProjectMembers(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))

	project, err := h.projectRepo.GetByID(c.Request.Context(), projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	members, err := h.projectRepo.GetProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve project members"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
	})
}

// DELETE /api/projects/:id/members/:userID
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requesterID := userIDVal.(int)

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	ctx := c.Request.Context()

	// Get the roles of both the requester and the target user
	requesterMember, err := h.projectRepo.GetProjectMember(ctx, projectID, requesterID)
	if err != nil || requesterMember == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this project"})
		return
	}

	targetMember, err := h.projectRepo.GetProjectMember(ctx, projectID, targetUserID)
	if err != nil || targetMember == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a member of this project"})
		return
	}

	// Check if the target user is the owner; owners cannot be removed
	if targetMember.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot remove project owner"})
		return
	}

	// Verify that the requester has sufficient permissions to remove the target user
	switch requesterMember.Role {
	case "owner":
		// Can remove any member except the owner (already checked above)
	case "admin":
		if targetMember.Role != "member" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admins can only remove standard members"})
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions to remove members"})
		return
	}

	// Remove the target user from the project
	if err := h.projectRepo.RemoveProjectMember(ctx, projectID, targetUserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove project member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project member removed successfully"})
}

// POST /api/projects/:id/leave
func (h *ProjectHandler) LeaveProject(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	userID := userIDVal.(int)

	// Get the member record for the user in the project
	member, err := h.projectRepo.GetProjectMember(c.Request.Context(), projectID, userID)
	if err != nil || member == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this project"})
		return
	}

	// Check if the user is the owner; owners cannot remove themselves
	if member.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "project owners cannot remove themselves"})
		return
	}

	// Remove the user from the project
	if err := h.projectRepo.RemoveProjectMember(c.Request.Context(), projectID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove project member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project member removed successfully"})
}

type UpdateRoleDTO struct {
	Role string `json:"role" binding:"required,oneof=admin member"`
}

// PATCH /api/projects/:id/members/:userID/role
func (h *ProjectHandler) UpdateMemberRole(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requesterID := userIDVal.(int)

	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var dto UpdateRoleDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body, expected role to be 'admin' or 'member'"})
		return
	}

	ctx := c.Request.Context()

	// Check if the requester is the owner of the project
	requester, err := h.projectRepo.GetProjectMember(ctx, projectID, requesterID)
	if err != nil || requester == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this project"})
		return
	}
	if requester.Role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "only project owner can change member roles"})
		return
	}

	// Check if the target user exists in the project
	targetMember, err := h.projectRepo.GetProjectMember(ctx, projectID, targetUserID)
	if err != nil || targetMember == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user is not a member of this project"})
		return
	}

	// Prevent changing the role of the owner
	if targetMember.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot change owner's role"})
		return
	}

	// Update the role of the target user
	input := repository.UdpateProjectMemberRoleInput{
		ProjectID: projectID,
		UserID:    targetUserID,
		NewRole:   dto.Role,
	}

	if err := h.projectRepo.UpdateProjectMemberRole(ctx, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "member role updated successfully",
		"role":    dto.Role,
	})
}

type UpdateProjectDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// PUT /api/projects/:id
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var dto UpdateProjectDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: name (3-100 chars) and description (max 500 chars) are required"})
		return
	}

	ctx := c.Request.Context()

	project, err := h.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	input := repository.UpdateProjectInput{
		ProjectID:   projectID,
		Name:        dto.Name,
		Description: dto.Description,
	}

	if err := h.projectRepo.UpdateProject(ctx, input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update project"})
		return
	}

	updatedProject, err := h.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "project updated successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project updated successfully",
		"project": updatedProject,
	})
}

// DELETE /api/projects/:id
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	ctx := c.Request.Context()

	project, err := h.projectRepo.GetByID(ctx, projectID)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if err := h.projectRepo.DeleteProject(ctx, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted successfully"})
}

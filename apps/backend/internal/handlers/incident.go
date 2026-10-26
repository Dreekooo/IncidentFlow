package handlers

import (
	"net/http"
	"strconv"

	"incident-flow/backend/internal/models"
	"incident-flow/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	incidentRepo repository.IncidentRepository
}

func NewIncidentHandler(incidentRepo repository.IncidentRepository) *IncidentHandler {
	return &IncidentHandler{incidentRepo: incidentRepo}
}

type CreateIncidentRequest struct {
	Title       string                  `json:"title" binding:"required"`
	Description string                  `json:"description" binding:"required"`
	Priority    models.IncidentPriority `json:"priority" binding:"required"`
	Status      models.IncidentStatus   `json:"status" binding:"required"`
	AssignedTo  *int                    `json:"assigned_to"`
}

// POST /api/projects/:id/incidents
func (h *IncidentHandler) CreateIncident(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	incident := &models.Incident{
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      req.Status,
		AssignedTo:  req.AssignedTo,
	}

	err = h.incidentRepo.CreateIncident(c.Request.Context(), incident)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident"})
		return
	}

	c.JSON(http.StatusCreated, incident)
}

// GET /api/projects/:id/incidents/:incidentID
func (h *IncidentHandler) GetIncidentByID(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	incidentID, err := strconv.Atoi(c.Param("incidentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	incident, err := h.incidentRepo.GetIncidentByID(c.Request.Context(), projectID, incidentID)
	if err != nil {
		if err.Error() == "incident not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incident"})
		return
	}

	c.JSON(http.StatusOK, incident)
}

// GET /api/projects/:id/incidents
func (h *IncidentHandler) ListIncidentsByProjectID(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	incidents, err := h.incidentRepo.ListIncidentsByProjectID(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve incidents"})
		return
	}

	c.JSON(http.StatusOK, incidents)
}

// PUT /api/projects/:id/incidents/:incidentID
func (h *IncidentHandler) UpdateIncident(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	incidentID, err := strconv.Atoi(c.Param("incidentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	var req models.Incident
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = incidentID
	req.ProjectID = projectID

	err = h.incidentRepo.UpdateIncident(c.Request.Context(), &req)
	if err != nil {
		if err.Error() == "incident not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident updated successfully"})
}

// PUT /api/projects/:id/incidents/:incidentID/assign
func (h *IncidentHandler) AssignIncident(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	incidentID, err := strconv.Atoi(c.Param("incidentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	var req struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.incidentRepo.AssignIncident(c.Request.Context(), projectID, incidentID, req.UserID)
	if err != nil {
		if err.Error() == "incident not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident assigned successfully"})
}

// DELETE /api/projects/:id/incidents/:incidentID
func (h *IncidentHandler) DeleteIncident(c *gin.Context) {
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	incidentID, err := strconv.Atoi(c.Param("incidentID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid incident ID"})
		return
	}

	err = h.incidentRepo.DeleteIncident(c.Request.Context(), projectID, incidentID)
	if err != nil {
		if err.Error() == "incident not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Incident not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Incident deleted successfully"})
}

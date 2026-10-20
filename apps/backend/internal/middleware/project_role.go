package middleware

import (
	"incident-flow/backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RequireProjectRole(projectRepo repository.ProjectRepository, requiredRole string) gin.HandlerFunc {
	roleHierarchy := map[string]int{
		"member": 1,
		"admin":  2,
		"owner":  3,
	}

	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		userID := userIDVal.(int)

		projectID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
			c.Abort()
			return
		}

		ctx := c.Request.Context()

		// Check if the project exists (404)
		project, err := projectRepo.GetByID(ctx, projectID)
		if err != nil || project == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			c.Abort()
			return
		}

		// Check if the user is a member of the project (403)
		member, err := projectRepo.GetProjectMember(ctx, projectID, userID)
		if err != nil || member == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not a member of this project"})
			c.Abort()
			return
		}

		// Check if the user's role meets the required role level (403)
		userRoleLevel := roleHierarchy[string(member.Role)]
		requiredRoleLevel := roleHierarchy[requiredRole]

		if userRoleLevel < requiredRoleLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient project permissions"})
			c.Abort()
			return
		}

		c.Set("projectRole", member.Role)
		c.Next()
	}
}

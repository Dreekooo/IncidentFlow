package server

import (
	"incident-flow/backend/internal/handlers"
	"incident-flow/backend/internal/middleware"
	"net/http"

	"incident-flow/backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authHandler *handlers.AuthHandler,
	projectHandler *handlers.ProjectHandler,
	incidentHandler *handlers.IncidentHandler,
	projectRepo repository.ProjectRepository,
) *gin.Engine {
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

		projects := protected.Group("/projects")
		{
			projects.POST("", projectHandler.CreateProject)    // POST /api/projects
			projects.POST("/join", projectHandler.JoinProject) // POST /api/projects/join

			projects.GET("/:id", middleware.RequireProjectRole(projectRepo, "member"), projectHandler.GetProjectByID)            // GET /api/projects/:id
			projects.GET("/:id/join-code", middleware.RequireProjectRole(projectRepo, "member"), projectHandler.GetJoinCode)     // GET /api/projects/:id/join-code
			projects.GET("/:id/members", middleware.RequireProjectRole(projectRepo, "member"), projectHandler.GetProjectMembers) // GET /api/projects/:id/members
			projects.POST("/:id/leave", middleware.RequireProjectRole(projectRepo, "member"), projectHandler.LeaveProject)       // POST /api/projects/:id/leave

			projects.GET("/:id/requests", middleware.RequireProjectRole(projectRepo, "admin"), projectHandler.GetJoinRequests)                // GET /api/projects/:id/requests
			projects.POST("/:id/requests/handle", middleware.RequireProjectRole(projectRepo, "admin"), projectHandler.HandleJoinRequest)      // POST /api/projects/:id/requests/handle
			projects.PATCH("/:id/members/:userID/role", middleware.RequireProjectRole(projectRepo, "admin"), projectHandler.UpdateMemberRole) // PATCH /api/projects/:id/members/:userID/role
			projects.DELETE("/:id/members/:userID", middleware.RequireProjectRole(projectRepo, "admin"), projectHandler.RemoveProjectMember)  // DELETE /api/projects/:id/members/:userID

			projects.PUT("/:id", middleware.RequireProjectRole(projectRepo, "owner"), projectHandler.UpdateProject)    // PUT /api/projects/:id
			projects.DELETE("/:id", middleware.RequireProjectRole(projectRepo, "owner"), projectHandler.DeleteProject) // DELETE /api/projects/:id

			projects.POST("/:id/incidents", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.CreateIncident)                   // POST /api/projects/:id/incidents
			projects.GET("/:id/incidents", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.ListIncidentsByProjectID)          // GET /api/projects/:id/incidents
			projects.GET("/:id/incidents/:incidentID", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.GetIncidentByID)       // GET /api/projects/:id/incidents/:incidentID
			projects.PUT("/:id/incidents/:incidentID", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.UpdateIncident)        // PUT /api/projects/:id/incidents/:incidentID
			projects.PUT("/:id/incidents/:incidentID/assign", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.AssignIncident) // PUT /api/projects/:id/incidents/:incidentID/assign
			projects.DELETE("/:id/incidents/:incidentID", middleware.RequireProjectRole(projectRepo, "member"), incidentHandler.DeleteIncident)     // DELETE /api/projects/:id/incidents/:incidentID
		}
	}

	return r
}

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

			// Actions for project members (Member level)
			projectMember := projects.Group("/:id")
			projectMember.Use(middleware.RequireProjectRole(projectRepo, "member"))
			{
				projectMember.GET("", projectHandler.GetProjectByID)            // GET /api/projects/:id
				projectMember.GET("/join-code", projectHandler.GetJoinCode)     // GET /api/projects/:id/join-code
				projectMember.GET("/members", projectHandler.GetProjectMembers) // GET /api/projects/:id/members
				projectMember.POST("/leave", projectHandler.LeaveProject)       // POST /api/projects/:id/leave

				// View incidents for project members (inherits the "member" role)
				incidents := projectMember.Group("/incidents")
				{
					incidents.GET("", incidentHandler.ListIncidentsByProjectID)    // GET /api/projects/:id/incidents
					incidents.GET("/:incidentID", incidentHandler.GetIncidentByID) // GET /api/projects/:id/incidents/:incidentID
				}
			}

			// Actions for project management (Admin level)
			projectAdmin := projects.Group("/:id")
			projectAdmin.Use(middleware.RequireProjectRole(projectRepo, "admin"))
			{
				projectAdmin.GET("/requests", projectHandler.GetJoinRequests)                // GET /api/projects/:id/requests
				projectAdmin.POST("/requests/handle", projectHandler.HandleJoinRequest)      // POST /api/projects/:id/requests/handle
				projectAdmin.PATCH("/members/:userID/role", projectHandler.UpdateMemberRole) // PATCH /api/projects/:id/members/:userID/role
				projectAdmin.DELETE("/members/:userID", projectHandler.RemoveProjectMember)  // DELETE /api/projects/:id/members/:userID

				// Managing incidents for project admins (inherits the "admin" role)
				incidents := projectMember.Group("/incidents")
				{
					incidents.POST("", incidentHandler.CreateIncident)                   // POST /api/projects/:id/incidents
					incidents.PUT("/:incidentID", incidentHandler.UpdateIncident)        // PUT /api/projects/:id/incidents/:incidentID
					incidents.PUT("/:incidentID/assign", incidentHandler.AssignIncident) // PUT /api/projects/:id/incidents/:incidentID/assign
					incidents.DELETE("/:incidentID", incidentHandler.DeleteIncident)     // DELETE /api/projects/:id/incidents/:incidentID
				}
			}

			// Actions for critical project actions (Owner level)
			projectOwner := projects.Group("/:id")
			projectOwner.Use(middleware.RequireProjectRole(projectRepo, "owner"))
			{
				projectOwner.PUT("", projectHandler.UpdateProject)    // PUT /api/projects/:id
				projectOwner.DELETE("", projectHandler.DeleteProject) // DELETE /api/projects/:id
			}
		}
	}

	return r
}

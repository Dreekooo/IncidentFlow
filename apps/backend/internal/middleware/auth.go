package middleware

import (
	"incident-flow/backend/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from Authorization header
// Expected header: Authorization: Bearer <token>
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			c.Abort()
			return
		}

		tokenString, ok := ExtractBearerToken(authHeader)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		// Parse and validate token
		claims, err := auth.ParseAuthToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Store claims in context for handler to access
		c.Set("userID", claims.UserID)
		c.Set("claims", claims)

		// Continue to next handler
		c.Next()
	}
}

func ExtractBearerToken(authHeader string) (string, bool) {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", false
	}

	return parts[1], true
}

// GetUserIDFromContext extracts userID from context (set by AuthMiddleware)
func GetUserIDFromContext(c *gin.Context) int {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}
	id, ok := userID.(int)
	if !ok {
		return 0
	}
	return id
}

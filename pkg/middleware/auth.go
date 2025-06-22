package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/divinecoid/oneagent/internal/model"
	"github.com/divinecoid/oneagent/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No session ID provided",
			})
			return
		}

		session, err := m.authService.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired session",
			})
			return
		}

		// Parse session data
		var sessionData map[string]interface{}
		if err := json.Unmarshal(session.Data, &sessionData); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to parse session data",
			})
			return
		}

		// Set session data in context
		c.Set("session", session)
		c.Set("user_id", int64(sessionData["user_id"].(float64)))
		c.Set("user_email", sessionData["email"])
		c.Set("user_role", sessionData["role"])

		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...model.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "No role information found",
			})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid role format",
			})
			return
		}

		currentRole := model.Role(roleStr)
		allowed := false

		// Super admin has access to everything
		if currentRole == model.RoleSuperAdmin {
			allowed = true
		} else {
			// Check if user's role matches any of the required roles
			for _, role := range roles {
				if currentRole == role {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		c.Next()
	}
}
package middleware

import (
	"rvcs/internal/api/response"
	"rvcs/internal/security"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	UserContextKey   = "user_id"
	RoleContextKey   = "role"
	DeviceContextKey = "device_id"
)

// AuthMiddleware verifies web user JWT.
func AuthMiddleware(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		c.Set(UserContextKey, claims.UserID)
		c.Set(RoleContextKey, claims.Role)
		c.Next()
	}
}

// DeviceAuthMiddleware verifies device JWT.
func DeviceAuthMiddleware(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid device token")
			c.Abort()
			return
		}

		if claims.DeviceID == "" {
			response.Unauthorized(c, "invalid device token")
			c.Abort()
			return
		}

		c.Set(DeviceContextKey, claims.DeviceID)
		c.Next()
	}
}

// AdminOnlyMiddleware allows admin-only access.
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(RoleContextKey)
		if !exists {
			response.Forbidden(c, "role not found in context")
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			response.Forbidden(c, "invalid role type")
			c.Abort()
			return
		}

		if roleStr != "admin" {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

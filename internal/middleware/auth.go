package middleware

import (
	"net/http"
	"os"
	"pvz-service/internal/api"
	"pvz-service/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var jwtAuth *auth.JWTAuth

func InitJWT() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret-key-change-in-production"
	}
	jwtAuth = auth.NewJWTAuth(secret)
}

// GenerateToken - генерация токена
func GenerateToken(userID uuid.UUID, role api.UserRole) (string, error) {
	return jwtAuth.GenerateToken(userID, role)
}

// AuthMiddleware - проверка токена
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := jwtAuth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireModerator - проверка роли модератора
func RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != api.UserRoleModerator {
			c.JSON(http.StatusForbidden, gin.H{"message": "moderator access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireEmployee - проверка роли сотрудника
func RequireEmployee() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || (role != api.UserRoleEmployee && role != api.UserRoleModerator) {
			c.JSON(http.StatusForbidden, gin.H{"message": "employee access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserIDFromContext - получить userID из контекста
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := userID.(uuid.UUID)
	return id, ok
}

// GetRoleFromContext - получить роль из контекста
func GetRoleFromContext(c *gin.Context) (api.UserRole, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	r, ok := role.(api.UserRole)
	return r, ok
}

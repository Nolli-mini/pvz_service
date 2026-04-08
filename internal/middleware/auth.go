package middleware

import (
	"net/http"
	"os"
	"pvz-service/internal/api"
	"pvz-service/internal/auth"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	jwtAuth     *auth.JWTAuth
	jwtAuthOnce sync.Once
)

// initJWTAuth инициализирует JWT аутентификацию (ленивая инициализация)
func initJWTAuth() {
	jwtAuthOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "test-secret-key-for-testing"
		}
		jwtAuth = auth.NewJWTAuth(secret)
	})
}

// GenerateToken генерирует JWT токен
func GenerateToken(userID uuid.UUID, role api.UserRole) (string, error) {
	initJWTAuth()
	return jwtAuth.GenerateToken(userID, role)
}

// AuthMiddleware проверяет токен
func AuthMiddleware() gin.HandlerFunc {
	initJWTAuth()
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

// RequireModerator проверяет роль модератора
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

// RequireEmployee проверяет роль сотрудника
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

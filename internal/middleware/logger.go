package middleware

import (
	"pvz-service/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Log.WithFields(map[string]interface{}{
			"method":  method,
			"path":    path,
			"status":  statusCode,
			"latency": latency.Milliseconds(),
			"ip":      c.ClientIP(),
		}).Info("")

		if statusCode >= 400 {
			logger.Log.WithFields(map[string]interface{}{
				"method": method,
				"path":   path,
				"status": statusCode,
			}).Warn("")
		}
	}
}

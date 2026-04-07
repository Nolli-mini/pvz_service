package middleware

import (
	"pvz-service/internal/metrics"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware собирает метрики для каждого запроса
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Обрабатываем запрос
		c.Next()

		// Собираем метрики после обработки
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		// Определяем endpoint (заменяем динамические ID на шаблон)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Увеличиваем счетчик запросов
		metrics.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			endpoint,
			string(rune(status)),
		).Inc()

		// Записываем длительность запроса
		metrics.HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			endpoint,
		).Observe(duration)
	}
}

package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()

		requestID, _ := c.Get(RequestIDKey)
		logger.Info("http_request",
			slog.String("request_id", toString(requestID)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(startedAt)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func toString(value any) string {
	if raw, ok := value.(string); ok {
		return raw
	}
	return ""
}

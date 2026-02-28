package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		uri := c.Request.RequestURI
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)

		status := c.Writer.Status()
		size := c.Writer.Size()

		logger.Info("Request processed",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.Int("status", status),
			zap.Int("size", size),
		)
	}
}

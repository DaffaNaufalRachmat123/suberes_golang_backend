package middleware

import (
	"time"

	"suberes_golang/logger"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware logs every HTTP request as a structured zerolog event.
// Requests that result in 5xx are logged at Error level, 4xx at Warn, rest at Info.
// Must be registered AFTER RequestIDMiddleware so the request ID is available.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		requestID, _ := c.Get("request_id")

		if raw != "" {
			path = path + "?" + raw
		}

		event := logger.Logger.Info()
		if statusCode >= 500 {
			event = logger.Logger.Error()
		} else if statusCode >= 400 {
			event = logger.Logger.Warn()
		}

		event.
			Str("request_id", requestID.(string)).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", statusCode).
			Str("ip", c.ClientIP()).
			Dur("latency_ms", latency).
			Str("user_agent", c.Request.UserAgent()).
			Int("bytes_out", c.Writer.Size()).
			Msg("http_request")
	}
}

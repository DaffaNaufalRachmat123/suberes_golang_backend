package middleware

import (
	"fmt"
	"net/http"

	"suberes_golang/logger"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware recovers from panics, logs the stack trace with zerolog,
// reports the event to Sentry, and returns a 500 JSON response.
// Use this instead of gin's built-in recovery so panics are observable in Sentry.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				// Report to Sentry — recover provides the panic value and stack.
				sentry.CurrentHub().RecoverWithContext(c.Request.Context(), err)

				logger.Logger.Error().
					Str("request_id", fmt.Sprintf("%v", requestID)).
					Str("method", c.Request.Method).
					Str("path", c.Request.URL.Path).
					Str("ip", c.ClientIP()).
					Interface("panic", err).
					Msg("panic_recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"server_message": "Internal server error",
					"status":         "failure",
				})
			}
		}()
		c.Next()
	}
}

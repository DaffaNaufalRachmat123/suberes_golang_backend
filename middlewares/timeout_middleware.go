package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware injects a request-scoped deadline into the request context.
// Any downstream operation that respects context cancellation (GORM queries,
// HTTP client calls, Redis commands) will automatically be cancelled when the
// timeout elapses. Default recommendation: 30 s.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

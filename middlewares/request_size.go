package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize limits the request body size.
// Default: 2MB for JSON endpoints. File upload endpoints should override with higher limits.
const DefaultMaxBodySize = 2 << 20 // 2MB

func RequestSizeLimiter(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body too large",
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

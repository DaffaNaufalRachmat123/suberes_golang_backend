package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestIDMiddleware ensures every request carries a unique X-Request-ID.
// If the client already sends one, it is kept; otherwise a UUID v4 is generated.
// The ID is written back to the response header so clients can correlate logs.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header(RequestIDHeader, requestID)
		c.Next()
	}
}

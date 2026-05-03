package middleware

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// XenditCallbackTokenMiddleware validates the Xendit callback token.
func XenditCallbackTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		xenditCallbackToken := os.Getenv("XENDIT_VERIFICATION_TOKEN")
		if xenditCallbackToken == "" {
			log.Println("XENDIT_VERIFICATION_TOKEN environment variable not set")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if c.GetHeader("x-callback-token") != xenditCallbackToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid callback token"})
			return
		}

		c.Next()
	}
}

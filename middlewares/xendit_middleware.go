package middleware

import (
	"net/http"
	"os"

	"suberes_golang/i18n"

	"github.com/gin-gonic/gin"
)

// XenditCallbackTokenMiddleware validates the Xendit callback token.
func XenditCallbackTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		xenditCallbackToken := os.Getenv("XENDIT_VERIFICATION_TOKEN")
		if xenditCallbackToken == "" {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if c.GetHeader("x-callback-token") != xenditCallbackToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgInvalidCallbackTok),
				"status":         "failure",
			})
			return
		}

		c.Next()
	}
}

package middleware

import (
	"fmt"
	"net/http"
	"suberes_golang/models"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware adalah pengganti role_authenticator
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ambil user dari context (yang diset di AuthMiddleware)
		userCtx, exists := c.Get("currentUser")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "Unauthorized", "status": "failure"})
			return
		}

		user := userCtx.(models.User) // Type assertion ke struct User

		fmt.Printf("User Type : %s\n", user.UserType)

		isFound := false
		for _, role := range allowedRoles {
			if role == user.UserType {
				isFound = true
				break
			}
		}

		if !isFound {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Unauthorized",
				"status":         "failure",
			})
			return
		}

		c.Next()
	}
}

package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"suberes_golang/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type JWTClaims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		for k, v := range c.Request.Header {
			fmt.Println(k, v)
		}
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "UNAUTHORIZED", "status": "failure"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "Invalid token format", "status": "failure"})
			return
		}

		secretKey := os.Getenv("SECRET_KEY")
		if secretKey == "" {
			secretKey = "SuberesIndustries"
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "Invalid or expired token", "status": "failure"})
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "Invalid token claims", "status": "failure"})
			return
		}

		fmt.Printf("JWT Verify ID : %s\n", claims.ID)

		var user models.User

		err = db.Select("id, complete_name, email, phone_number, country_code, user_type, user_gender, is_logged_in, user_status, is_active, is_mitra_accepted, is_mitra_activated, is_suspended").
			Where("id = ?", claims.ID).
			First(&user).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "User not found", "status": "failure"})
			return
		}

		isValid := false

		if user.UserType == "admin" || user.UserType == "superadmin" {
			if user.IsActive == "yes" {
				isValid = true
			}
		} else if user.UserType == "mitra" {
			if user.IsSuspended != "1" && user.IsMitraAccepted == "1" && user.IsMitraActivated == "1" {
				isValid = true
			}
		} else {
			isValid = true
		}

		if !isValid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": "Account is not active or authorized", "status": "failure"})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

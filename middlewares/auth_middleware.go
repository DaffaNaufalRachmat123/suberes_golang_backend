package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"suberes_golang/i18n"
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgUnauthorized), "status": "failure"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgTokenInvalidFormat), "status": "failure"})
			return
		}

		secretKey := os.Getenv("SECRET_KEY")
		if secretKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"server_message": i18n.Tc(c, i18n.MsgServerConfigError), "status": "failure"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgTokenExpired), "status": "failure", "need_refresh": true})
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgTokenInvalidOrExpired), "status": "failure"})
			}
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgTokenInvalidClaims), "status": "failure"})
			return
		}

		var user models.User

		err = db.Select("id, complete_name, email, phone_number, country_code, user_type, user_gender, is_logged_in, user_status, is_active, is_mitra_accepted, is_mitra_activated, is_suspended, device_id").
			Where("id = ?", claims.ID).
			First(&user).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgNotFound), "status": "failure"})
			return
		}

		// Validate device binding — if user has a bound device_id, incoming request must match
		if user.DeviceID != "" {
			deviceID := c.GetHeader("device_id")
			if deviceID == "" || deviceID != user.DeviceID {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"server_message": i18n.Tc(c, i18n.MsgDeviceMismatch),
					"status":         "failure",
				})
				return
			}
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(c, i18n.MsgAccountNotActive), "status": "failure"})
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

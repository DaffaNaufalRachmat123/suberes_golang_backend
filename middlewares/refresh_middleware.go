package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"suberes_golang/i18n"
	"suberes_golang/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func RefreshTokenMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgMissingRefreshToken),
				"status":         "failure",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgTokenInvalidFormat),
				"status":         "failure",
			})
			return
		}

		secretKey := os.Getenv("SECRET_KEY_REFRESH")
		if secretKey == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"server_message": i18n.Tc(c, i18n.MsgServerConfigError), "status": "failure"})
			return
		}
		// ✅ parse tanpa validasi expiry
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(secretKey), nil
		})

		if err != nil || token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgRefreshTokenInvalid),
				"status":         "failure",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgTokenInvalidClaims),
				"status":         "failure",
			})
			return
		}

		idVal, ok := claims["id"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgTokenInvalidPayload),
				"status":         "failure",
			})
			return
		}
		userID := fmt.Sprintf("%v", idVal)

		// ✅ hash token
		hash := sha256.Sum256([]byte(tokenString))
		tokenHash := hex.EncodeToString(hash[:])

		var stored models.RefreshToken
		err = db.Where("token_hash = ? AND users_id = ? AND revoked = ?", tokenHash, userID, "0").
			First(&stored).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgRefreshTokenNotRecognized),
				"status":         "failure",
			})
			return
		}

		if time.Now().After(stored.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": i18n.Tc(c, i18n.MsgRefreshTokenExpired),
				"status":         "failure",
			})
			return
		}

		// Validate device_id if present in the stored token
		if stored.DeviceID != "" {
			deviceID := c.GetHeader("device_id")
			if deviceID == "" || deviceID != stored.DeviceID {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"server_message": i18n.Tc(c, i18n.MsgDeviceMismatch),
					"status":         "failure",
				})
				return
			}
		}

		// ✅ inject ke context
		c.Set("refreshUserID", userID)
		c.Set("refreshTokenRecord", stored)

		c.Next()
	}
}

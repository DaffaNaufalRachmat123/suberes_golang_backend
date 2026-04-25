package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

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
				"server_message": "Missing refresh token",
				"status":         "failure",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Invalid token format",
				"status":         "failure",
			})
			return
		}

		secretKey := os.Getenv("SECRET_KEY_REFRESH")
		if secretKey == "" {
			secretKey = "SuberesIndustries"
		}
		// ✅ parse tanpa validasi expiry
		parser := jwt.NewParser(jwt.WithoutClaimsValidation())
		token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method")
			}
			return []byte(secretKey), nil
		})

		fmt.Printf("Token : %s", token)
		fmt.Println(err)

		if err != nil || token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Invalid refresh token",
				"status":         "failure",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Invalid claims",
				"status":         "failure",
			})
			return
		}

		idVal, ok := claims["id"]
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Invalid token payload",
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
				"server_message": "Refresh token not recognized",
				"status":         "failure",
			})
			return
		}

		if time.Now().After(stored.ExpiresAt) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"server_message": "Refresh token expired",
				"status":         "failure",
			})
			return
		}

		// ✅ inject ke context
		c.Set("refreshUserID", userID)
		c.Set("refreshTokenRecord", stored)

		c.Next()
	}
}

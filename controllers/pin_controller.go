package controllers

import (
	"errors"
	"net/http"
	"suberes_golang/models"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PinController struct {
	PinService *services.PinService
}

// GetPublicKey GET /api/pins/customer/public_key
func (c *PinController) GetPublicKey(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	keys, err := c.PinService.GetPublicKeys(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, keys)
}

// GetPinStatus GET /api/pins/customer/pin_status
func (c *PinController) GetPinStatus(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	status, err := c.PinService.GetPinStatus(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, status)
}

// PinCheck POST /api/pins/customer/pin_check
func (c *PinController) PinCheck(ctx *gin.Context) {
	var body struct {
		PinType string `json:"pin_type" binding:"required,oneof=pay disbursement"`
		Pin     string `json:"pin" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	if err := c.PinService.CheckPin(user.ID, body.PinType, body.Pin); err != nil {
		if err.Error() == "old PIN is different" {
			ctx.JSON(http.StatusForbidden, gin.H{"server_message": "old PIN is different", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "pin check succeeded", "status": "success"})
}

// RequestChangePin POST /api/pins/customer/request_change_pin
func (c *PinController) RequestChangePin(ctx *gin.Context) {
	var body struct {
		PinType string `json:"pin_type" binding:"required,oneof=pay disbursement"`
		Pin     string `json:"pin" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	otpRecord, err := c.PinService.RequestChangePin(user.ID, body.PinType, body.Pin)
	if err != nil {
		if err.Error() == "old PIN is different" {
			ctx.JSON(http.StatusForbidden, gin.H{"server_message": "old PIN is different", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "otp number sent",
		"otp_timeout":    ctx.GetString("OTP_TIMEOUT"),
		"status":         "success",
		"user_otp_id":    otpRecord.ID,
	})
}

// OtpValidate POST /api/pins/customer/otp_validate
func (c *PinController) OtpValidate(ctx *gin.Context) {
	var body struct {
		PinType   string `json:"pin_type" binding:"required,oneof=pay disbursement"`
		UserOtpID int    `json:"user_otp_id" binding:"required"`
		OtpCode   string `json:"otp_code" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	if err := c.PinService.ValidateOtp(user.ID, body.PinType, body.UserOtpID, body.OtpCode); err != nil {
		if err.Error() == "OTP Code is wrong" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": "OTP Code is wrong", "status": "failure"})
			return
		}
		if err.Error() == "invalid request" || errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid request", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "otp valid",
		"status":         "success",
		"user_otp_id":    body.UserOtpID,
	})
}

// ConfigurePin POST /api/pins/customer/configure/pin
func (c *PinController) ConfigurePin(ctx *gin.Context) {
	var body struct {
		PinType   string `json:"pin_type" binding:"required,oneof=pay disbursement"`
		Pin       string `json:"pin" binding:"required"`
		ChangePin string `json:"change_pin" binding:"required,oneof=0 1"`
		UserOtpID *int   `json:"user_otp_id"`
		OtpCode   string `json:"otp_code"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}
	userOtpID := 0
	if body.UserOtpID != nil {
		userOtpID = *body.UserOtpID
	}
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)
	pinStatus, err := c.PinService.ConfigurePin(user.ID, body.PinType, body.ChangePin, body.Pin, userOtpID, body.OtpCode)
	if err != nil {
		if err.Error() == "invalid request" {
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid request", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	serverMessage := "pin successfully configured"
	if body.ChangePin == "1" {
		serverMessage = "pin successfully changed"
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": serverMessage,
		"status":         "success",
		"pin_status":     pinStatus,
	})
}

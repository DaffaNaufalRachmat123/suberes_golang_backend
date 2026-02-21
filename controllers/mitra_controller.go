package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/services"
	"time"

	"github.com/gin-gonic/gin"
)

type MitraController struct {
	MitraService *services.MitraService
}

func NewMitraController(mitraService *services.MitraService) *MitraController {
	return &MitraController{
		MitraService: mitraService,
	}
}

func (c *MitraController) Login(ctx *gin.Context) {
	var loginDTO dtos.MitraLoginDTO
	if err := ctx.ShouldBindJSON(&loginDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.MitraService.Login(loginDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) Register(ctx *gin.Context) {
	var registerDTO dtos.MitraRegisterDTO
	if err := ctx.ShouldBind(&registerDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File
	filePaths := make(map[string]string)

	for key, fileHeaders := range files {
		if len(fileHeaders) > 0 {
			file := fileHeaders[0]
			date := time.Now()
			dateImage := fmt.Sprintf("%d-%d-%d_%d-%d-%d", date.Year(), date.Month(), date.Day(), date.Hour(), date.Minute(), date.Second())
			filename := fmt.Sprintf("MITRA_CANDIDATE_IMAGE%s_%s", dateImage, file.Filename)
			path := filepath.Join("images/mitra_candidate", filename)

			if err := ctx.SaveUploadedFile(file, path); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file: %s", err.Error())})
				return
			}
			filePaths[key] = path
		}
	}

	createdMitra, err := c.MitraService.Register(registerDTO, filePaths)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "register successful",
		"status":         "ok",
		"data":           createdMitra,
	})
}

func (c *MitraController) Profile(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	timezoneCode := ctx.Query("timezone_code")

	response, err := c.MitraService.GetProfile(mitraID, timezoneCode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) GetEmailPassword(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	response, err := c.MitraService.GetEmailPassword(mitraID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) ChangePassword(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	var changePasswordDTO dtos.ChangePasswordDTO
	if err := ctx.ShouldBindJSON(&changePasswordDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.MitraService.ChangePassword(mitraID, changePasswordDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "change password successfuly", "status": "success"})
}

func (c *MitraController) UpdateMitraStatus(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	status := ctx.Param("status")
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"server_message": "Customer not found",
			"status":         "failure",
		})
	}
	user := userCtx.(models.User)
	var req dtos.SuspendRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failed",
		})
		return
	}
	code, err := c.MitraService.UpdateMitraStatus(ctx, mitraID, status, user.UserType, req.SuspendedReason)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Mitra status updated",
		"status":         "success",
	})
}

func (c *MitraController) ChangeEmail(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	var changeEmailDTO dtos.ChangeEmailDTO
	if err := ctx.ShouldBindJSON(&changeEmailDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otpTimeout, err := c.MitraService.ChangeEmail(mitraID, changeEmailDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "email & password updatead",
		"status":         "success",
		"otp_timeout":    otpTimeout,
	})
}

func (c *MitraController) ChangeForgotPassword(ctx *gin.Context) {
	var forgotPasswordDTO dtos.ForgotPasswordDTO
	if err := ctx.ShouldBindJSON(&forgotPasswordDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.MitraService.ChangeForgotPassword(forgotPasswordDTO)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "Password updated", "status": "success"})
}

func (c *MitraController) RequestForgotPassword(ctx *gin.Context) {
	email := ctx.Param("email")
	otpTimeout, err := c.MitraService.RequestForgotPassword(email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "forgot password request succeeded",
		"status":         "success",
		"otp_timeout":    otpTimeout,
	})
}

func (c *MitraController) OTPValidatorForgotPassword(ctx *gin.Context) {
	var dto dtos.OTPValidatorForgotPasswordDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.MitraService.OTPValidatorForgotPassword(dto)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "email changed", "status": "success"})
}

func (c *MitraController) Logout(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	err := c.MitraService.Logout(mitraID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "logout success", "status": "success"})
}

func (c *MitraController) UpdateFirebaseToken(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	token := ctx.Param("firebase_token")
	err := c.MitraService.UpdateFirebaseToken(mitraID, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "refresh token updated", "status": "success"})
}

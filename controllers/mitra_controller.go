package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/i18n"
	"suberes_golang/models"
	"suberes_golang/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
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

	// Check account lockout
	if helpers.IsAccountLocked("mitra", loginDTO.Email) {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "account temporarily locked due to too many failed attempts", "failure_type": "ACCOUNT_LOCKED"})
		return
	}

	response, err := c.MitraService.Login(loginDTO)
	if err != nil {
		helpers.WriteAuditLog(helpers.AuditLog{
			Event:     helpers.AuditLoginFailed,
			IP:        ctx.ClientIP(),
			UserAgent: ctx.GetHeader("User-Agent"),
			UserType:  "mitra",
			Resource:  "/mitra/login",
			Details:   err.Error(),
			Success:   false,
		})
		helpers.RecordFailedLogin("mitra", loginDTO.Email)
		switch err.Error() {
		case "mitra not found":
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error(), "failure_type": "MITRA_NOT_FOUND"})
		case "password not match":
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error(), "failure_type": "PASSWORD_DOESNT_MATCH"})
		case "mitra already logged in another device, please log out first":
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error(), "failure_type": "MITRA_ALREADY_LOGGED_IN"})
		case "mitra not activated":
			ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "failure_type": "MITRA_NOT_ACTIVATED"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	helpers.ClearFailedLogin("mitra", loginDTO.Email)
	helpers.WriteAuditLog(helpers.AuditLog{
		Event:     helpers.AuditLogin,
		IP:        ctx.ClientIP(),
		UserAgent: ctx.GetHeader("User-Agent"),
		UserType:  "mitra",
		Resource:  "/mitra/login",
		Success:   true,
	})

	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) RefreshToken(ctx *gin.Context) {
	userID := ctx.GetString("refreshUserID")

	tokenRecordRaw, exists := ctx.Get("refreshTokenRecord")
	if !exists {
		helpers.APIErrorResponse(ctx, "invalid token context", http.StatusUnauthorized)
		return
	}

	tokenRecord := tokenRecordRaw.(models.RefreshToken)

	result, err := c.MitraService.RefreshToken(userID, tokenRecord)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusUnauthorized)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (c *MitraController) Register(ctx *gin.Context) {

	var registerDTO dtos.MitraRegisterDTO

	// ambil json_data dari form
	jsonData := ctx.PostForm("json_data")

	if jsonData == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "json_data is required",
		})
		return
	}

	// parse json string -> struct
	if err := json.Unmarshal([]byte(jsonData), &registerDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid json_data format",
		})
		return
	}

	// validasi struct (optional kalau pakai binding tag)
	if err := validator.New().Struct(registerDTO); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// ambil file multipart
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid form data",
		})
		return
	}

	files := form.File
	filePaths := make(map[string]string)

	for key, fileHeaders := range files {
		if len(fileHeaders) > 0 {

			file := fileHeaders[0]

			if err := helpers.ValidateUploadedFile(file); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			date := time.Now()
			dateImage := fmt.Sprintf(
				"%d-%d-%d_%d-%d-%d",
				date.Year(),
				date.Month(),
				date.Day(),
				date.Hour(),
				date.Minute(),
				date.Second(),
			)

			filename := fmt.Sprintf(
				"MITRA_CANDIDATE_IMAGE%s_%s",
				dateImage,
				helpers.SanitizeFilename(file.Filename),
			)

			path := filepath.Join("images/mitra_candidate", filename)

			if err := ctx.SaveUploadedFile(file, path); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to save file: %s", err.Error()),
				})
				return
			}

			filePaths[key] = path
		}
	}

	createdMitra, err := c.MitraService.Register(registerDTO, filePaths)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message":        i18n.Tc(ctx, i18n.MsgRegisterSuccess),
		"status":                "ok",
		"register_success_text": "Pendaftaran mitra baru berhasil selanjutnya kita akan ngasih tau kamu lewat email untuk proses verifikasi data diri dan kelengkapan",
		"data":                  createdMitra,
	})
}

func (c *MitraController) Profile(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	timezoneCode := ctx.Query("timezone_code")

	response, err := c.MitraService.GetProfile(mitraID, timezoneCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "mitra not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) ShowPhone(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	mitra, err := c.MitraService.ShowPhone(mitraID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "mitra not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, mitra)
}

func (c *MitraController) Saldo(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	response, err := c.MitraService.GetSaldo(mitraID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "mitra not found"})
			return
		}
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
		if appErr, ok := err.(*helpers.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{"error": appErr.Message})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgPasswordChanged), "status": "success"})
}

func (c *MitraController) UpdateMitraStatus(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	status := ctx.Param("status")
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgCustomerNotFound),
			"status":         "failure",
		})
		return
	}
	user := userCtx.(models.User)
	suspendedReason := ""
	if status == "suspend" {
		var req dtos.SuspendRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"server_message": err.Error(),
				"status":         "failed",
			})
			return
		}
		suspendedReason = req.SuspendedReason
	}
	code, err := c.MitraService.UpdateMitraStatus(ctx, mitraID, status, user.UserType, suspendedReason)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraStatusUpdated),
		"status":         "success",
	})
}

func (c *MitraController) UpdateMitraActive(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	isActive := ctx.Param("isactive")
	code, err := c.MitraService.UpdateMitraActive(mitraID, isActive)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraActiveUpdated),
		"status":         "success",
	})
}

func (c *MitraController) UpdateMitraAutoBid(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	isAutoBid := ctx.Param("isautobid")
	code, err := c.MitraService.UpdateMitraAutoBid(mitraID, isAutoBid)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraAutoBidUpdated),
		"status":         "success",
	})
}

func (c *MitraController) UpdateMitraCoordinate(ctx *gin.Context) {
	mitraID := ctx.Param("id")

	var body struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		helpers.APIErrorResponse(ctx, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	code, err := c.MitraService.UpdateMitraCoordinate(mitraID, body.Latitude, body.Longitude)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraCoordUpdated),
		"status":         "success",
	})
}
func (c *MitraController) AdminIndex(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "5")
	search := ctx.DefaultQuery("search", "")

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	mitra, total, err := c.MitraService.AdminIndex(page, limit, search)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	response := helpers.GetPaginationData(ctx, mitra, len(mitra), 1, 5, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) GetMitraDetail(ctx *gin.Context) {

	idParam := ctx.Param("id")
	status := ctx.Param("status")
	timezone := ctx.DefaultQuery("timezone", "Asia/Jakarta")

	response, code, err := c.MitraService.GetMitraDetail(idParam, status, timezone, i18n.GetLang(ctx))
	if err != nil {
		ctx.JSON(code, gin.H{
			"server_message": err.Error(),
			"status":         "failed",
		})
		return
	}

	ctx.JSON(code, response)
}
func (c *MitraController) AdminUpdate(ctx *gin.Context) {
	var updateReq dtos.UpdateMitraRequest
	if err := ctx.ShouldBindJSON(&updateReq); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
	}
	code, err := c.MitraService.AdminUpdate(updateReq)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraUpdated),
		"status":         "success",
	})
}

func (c *MitraController) AdminCandidate(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "5"))

	search := ctx.Query("search")
	isExGolife := ctx.Query("is_ex_golife")
	kindOfMitra := ctx.Query("kind_of_mitra")

	data, total, err := c.MitraService.GetFilteredMitra(
		page,
		limit,
		search,
		isExGolife,
		kindOfMitra,
	)
	if err != nil {
		helpers.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	response := helpers.GetPaginationData(ctx, data, len(data), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *MitraController) UpdateMitraCandidate(ctx *gin.Context) {
	jsonData := ctx.PostForm("json_data")

	var req dtos.UpdateMitraCandidateRequest
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		ctx.JSON(400, gin.H{"message": "Invalid data request"})
		return
	}

	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(400, gin.H{"message": "Invalid mitra id"})
		return
	}

	basePath := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
		os.Getenv("MITRA_CANDIDATE_IMAGE_PATH"),
	)

	_ = os.MkdirAll(basePath, 0755)

	now := time.Now()

	savedFiles := []string{}
	filePayload := map[string]string{}

	handleUpload := func(field string, dbField string) error {
		fileHeader, err := ctx.FormFile(field)
		if err != nil {
			return nil
		}

		if err := helpers.ValidateUploadedFile(fileHeader); err != nil {
			return err
		}

		filename := fmt.Sprintf(
			"MITRA_CANDIDATE_IMG_%d-%02d-%02d_%02d-%02d-%02d_%s",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			helpers.SanitizeFilename(fileHeader.Filename),
		)

		fullPath := filepath.Join(basePath, filename)

		if err := ctx.SaveUploadedFile(fileHeader, fullPath); err != nil {
			return err
		}

		savedFiles = append(savedFiles, fullPath)
		filePayload[dbField] = os.Getenv("MITRA_CANDIDATE_IMAGE_PATH") + filename

		return nil
	}

	if err := handleUpload("ktp", "ktp_image"); err != nil {
		ctx.JSON(500, gin.H{"message": err.Error()})
		return
	}
	if err := handleUpload("kk", "kk_image"); err != nil {
		ctx.JSON(500, gin.H{"message": err.Error()})
		return
	}
	if err := handleUpload("profile_image", "user_profile_image"); err != nil {
		ctx.JSON(500, gin.H{"message": err.Error()})
		return
	}

	// 4️⃣ Call Service (NO ctx inside)
	err = c.MitraService.UpdateMitraCandidate(
		id,
		req,
		filePayload,
		savedFiles,
	)

	if err != nil {
		ctx.JSON(400, gin.H{"message": err.Error()})
		return
	}

	ctx.JSON(200, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgMitraCandidateUpdated),
		"status":         "success",
	})
}

func (c *MitraController) UpdateDocumentStatus(ctx *gin.Context) {
	var req dtos.DocumentStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	code, err := c.MitraService.UpdateDocumentStatus(req)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgDocumentStatusUpdated),
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
		"server_message": i18n.Tc(ctx, i18n.MsgEmailPasswordUpdated),
		"status":         "success",
		"otp_timeout":    otpTimeout,
	})
}

// OtpValidatorEmailVerification handles PUT /api/mitra/otp_validator/email_verification_code
func (c *MitraController) OtpValidatorEmailVerification(ctx *gin.Context) {
	var dto dtos.OtpValidatorEmailVerificationDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	newEmail, err := c.MitraService.OtpValidatorEmailVerification(dto)
	if err != nil {
		if appErr, ok := err.(*helpers.AppError); ok {
			ctx.JSON(appErr.Code, gin.H{
				"server_message": appErr.Message,
				"status":         "failure",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgEmailChanged),
		"status":         "success",
		"email":          newEmail,
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

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgPasswordUpdated), "status": "success"})
}

func (c *MitraController) RequestForgotPassword(ctx *gin.Context) {
	email := ctx.Param("email")
	otpTimeout, err := c.MitraService.RequestForgotPassword(email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgForgotPasswordSuccess),
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

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgEmailChanged), "status": "success"})
}

func (c *MitraController) Logout(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	err := c.MitraService.Logout(mitraID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgLogoutSuccess), "status": "success"})
}

func (c *MitraController) UpdateFirebaseToken(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	token := ctx.Param("firebase_token")
	err := c.MitraService.UpdateFirebaseToken(mitraID, token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgRefreshTokenUpdated), "status": "success"})
}

func (c *MitraController) InviteMitra(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	var req dtos.InviteMitraRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	code, err := c.MitraService.InviteMitra(mitraID, req.ScheduleID)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgMitraInvited), "status": "success"})
}

func (c *MitraController) TrainingStatus(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	status := ctx.Param("status")
	code, err := c.MitraService.TrainingStatus(mitraID, status)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgMitraStatusUpdated), "status": "success"})
}

func (c *MitraController) ActivateMitraStatus(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	status := ctx.Param("status")
	code, err := c.MitraService.ActivateMitraStatus(mitraID, status)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}
	var msgKey string
	if status == "successful" {
		msgKey = i18n.MsgMitraActivated
	} else {
		msgKey = i18n.MsgMitraRejected
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, msgKey), "update_status": status, "status": "success"})
}

func (c *MitraController) DashboardCount(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	result, err := c.MitraService.DashboardCount(mitraID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failed"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

// GetPendapatan handles GET /api/mitra/pendapatan/:mitra_id/:pendapatan_date
// pendapatan_date must be in "YYYY-MM-DD" format.
func (c *MitraController) GetPendapatan(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	pendapatanDate := ctx.Param("pendapatan_date")

	result, err := c.MitraService.GetPendapatan(mitraID, pendapatanDate)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "success", http.StatusOK, result)
}

// PhoneChange handles PUT /api/mitra/phone_change/:mitra_id/:phone_number
func (c *MitraController) PhoneChange(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	phoneNumber := ctx.Param("phone_number")

	otpTimeout, err := c.MitraService.PhoneChange(mitraID, phoneNumber)
	if err != nil {
		if err.Error() == "mitra data not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"server_message": i18n.Tc(ctx, i18n.MsgMitraDataNotFound),
				"status":         "failure",
			})
			return
		}
		if err.Error() == "phone number already used" {
			ctx.JSON(http.StatusConflict, gin.H{
				"server_message": i18n.Tc(ctx, i18n.MsgPhoneAlreadyUsed),
				"status":         "failure",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgOtpSent),
		"otp_timeout":    otpTimeout,
		"status":         "success",
	})
}

// OtpValidatorChangePhoneNumber handles PUT /api/mitra/otp_validator/change_phone_number
func (c *MitraController) OtpValidatorChangePhoneNumber(ctx *gin.Context) {
	var dto dtos.OtpValidatorChangePhoneDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	newPhone, err := c.MitraService.OtpValidatorChangePhoneNumber(dto)
	if err != nil {
		switch err.Error() {
		case "mitra not found":
			ctx.JSON(http.StatusNotFound, gin.H{
				"server_message": err.Error(),
				"status":         "failure",
			})
		case "Your OTP number not found":
			ctx.JSON(http.StatusBadRequest, gin.H{
				"server_message": err.Error(),
				"status":         "failure",
			})
		case "otp code is wrong":
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"server_message": err.Error(),
				"status":         "failure",
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"server_message": err.Error(),
				"status":         "failure",
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgPhoneChanged),
		"status":         "success",
		"phone_number":   newPhone,
	})
}

func (c *MitraController) UpdateRejectionOrderCount(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	mitraID := ctx.Param("mitra_id")

	code, err := c.MitraService.UpdateRejectionOrderCount(orderID, mitraID)
	if err != nil {
		helpers.JSONError(ctx, code, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgRejectionCountUpdated),
		"status":         "success",
	})
}

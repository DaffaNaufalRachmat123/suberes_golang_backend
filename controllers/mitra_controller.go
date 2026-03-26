package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
				file.Filename,
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

func (c *MitraController) UpdateMitraActive(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	isActive := ctx.Param("isactive")
	code, err := c.MitraService.UpdateMitraActive(mitraID, isActive)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Mitra active status updated",
		"status":         "success",
	})
}

func (c *MitraController) UpdateMitraAutoBid(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	isAutoBid := ctx.Param("isautobid")
	code, err := c.MitraService.UpdateMitraAutoBid(mitraID, isAutoBid)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Mitra auto bid status updated",
		"status":         "success",
	})
}

func (c *MitraController) UpdateMitraCoordinate(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	latitude := ctx.Param("latitude")
	longitude := ctx.Param("longitude")
	latitudeConvert, err := strconv.ParseFloat(latitude, 64)

	if err != nil {
		fmt.Println("Error during conversion:", err)
		helpers.APIErrorResponse(ctx, "Error during conversion : "+err.Error(), http.StatusInternalServerError)
		return
	}
	longitudeConvert, err := strconv.ParseFloat(longitude, 64)

	if err != nil {
		helpers.APIErrorResponse(ctx, "Error during conversion : "+err.Error(), http.StatusInternalServerError)
		return
	}
	code, err := c.MitraService.UpdateMitraCoordinate(mitraID, latitudeConvert, longitudeConvert)
	if err != nil {
		helpers.JSONError(ctx, code, err)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Mitra coordinate updated",
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

	response, code, err := c.MitraService.GetMitraDetail(idParam, status, timezone)
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
		"server_message": "Mitra updated",
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

		filename := fmt.Sprintf(
			"MITRA_CANDIDATE_IMG_%d-%02d-%02d_%02d-%02d-%02d_%s",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			fileHeader.Filename,
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
		"server_message": "mitra candidate data updated",
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
		"server_message": "Document status updated",
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

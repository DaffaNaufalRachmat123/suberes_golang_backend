package controllers

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	CustomerService *services.CustomerService
}

type OtpValidatorMailRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OtpCode string `json:"otp_code" binding:"required"`
}

type UpdateFirebaseTokenRequest struct {
	FirebaseToken string `json:"firebase_token" binding:"required"`
}

type LoginEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type RegisterRequest struct {
	CompleteName string `json:"complete_name"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	CountryCode  string `json:"country_code"`
	UserType     string `json:"user_type"`
}

type ChangePhoneMailRequest struct {
	ID          string `json:"id"`
	PhoneChange bool   `json:"phone_change"`
	MailChange  bool   `json:"mail_change"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

type OtpUpdatePhoneMailRequest struct {
	ID           string `json:"id"`
	CompleteName string `json:"complete_name"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	CountryCode  string `json:"country_code"`
	OtpCode      string `json:"otp_code"`
	PhoneChange  bool   `json:"phone_change"`
	MailChange   bool   `json:"mail_change"`
}

type UpdateUserRequest struct {
	CompleteName string `json:"complete_name"`
}

func (r *RegisterRequest) Validate() error {
	if r.CompleteName == "" {
		return errors.New("complete_name is required")
	}
	if r.Email == "" || !strings.Contains(r.Email, "@") {
		return errors.New("email is invalid")
	}
	if r.PhoneNumber == "" {
		return errors.New("phone_number is required")
	}
	if r.CountryCode == "" {
		return errors.New("country_code is required")
	}
	if r.UserType == "" {
		return errors.New("user_type is required")
	}
	return nil
}

func (c *CustomerController) ChangePhoneMail(ctx *gin.Context) {
	var req ChangePhoneMailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}

	// Memanggil Service
	resp, status, err := c.CustomerService.ChangePhoneMail(
		req.ID,
		req.PhoneChange,
		req.MailChange,
		req.PhoneNumber,
		req.Email,
	)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), status)
		return
	}

	// Menggunakan helper response sukses yang sudah dibuat sebelumnya
	// Karena format response JS sedikit custom (ada otp_timeout di root json),
	// kita bisa pakai ctx.JSON langsung atau sesuaikan helper.
	ctx.JSON(status, resp)
}

func (c *CustomerController) UserLogout(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"server_message": "Customer not found",
			"status":         "failure",
		})
	}
	user := userCtx.(models.User)
	resp, status, err := c.CustomerService.Logout(user.ID)
	if err != nil {
		ctx.JSON(status, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(status, resp)
}

func (c *CustomerController) GetCustomerProfile(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not found",
		})
	}
	user := userCtx.(models.User)
	resp, status, err := c.CustomerService.GetCustomerProfile(user.ID)
	if err != nil {
		ctx.JSON(status, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	helpers.APIResponse(ctx, "Customer found", status, resp)
}

func (c *CustomerController) UpdateFirebaseToken(ctx *gin.Context) {
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"server_message": "Customer not found",
			"status":         "failure",
		})
	}
	user := userCtx.(models.User)
	var req UpdateFirebaseTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	resp, status, err := c.CustomerService.UpdateFirebaseTokenCustomer(
		user.ID, req.FirebaseToken,
	)
	if err != nil {
		ctx.JSON(status, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(status, resp)
}

func (c *CustomerController) Register(ctx *gin.Context) {
	var req RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.CustomerService.Register(req.CompleteName, req.Email, req.PhoneNumber, req.CountryCode, "customer"); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusConflict)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "register successfully",
		"otp_timeout":    os.Getenv("OTP_TIMEOUT"),
		"status":         "success",
	})
}

func (c *CustomerController) LoginByEmail(ctx *gin.Context) {
	var req LoginEmailRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, err)
		return
	}

	_, err := c.CustomerService.LoginByEmail(req.Email)

	if err != nil {
		if err.Error() == "CUSTOMER_ALREADY_LOGGED_IN" {
			ctx.JSON(409, gin.H{
				"failure_type":   err.Error(),
				"server_message": "this account already logged in on other device",
				"status":         "failure",
			})
			return
		}

		if err.Error() == "CUSTOMER_NOT_FOUND" {
			ctx.JSON(404, gin.H{
				"failure_type":   err.Error(),
				"server_message": "email not available",
				"status":         "failure",
			})
			return
		}

		ctx.JSON(500, err)
		return
	}

	ctx.JSON(200, gin.H{
		"server_message": "otp number sent",
		"otp_timeout":    os.Getenv("OTP_TIMEOUT"),
		"status":         "success",
	})
}

func (c *CustomerController) UpdateUserProfile(ctx *gin.Context) {
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not found",
		})
	}
	userToken := userCtx.(models.User)
	resp, status, err := c.CustomerService.UpdateUserProfile(
		userToken.ID,
		req.CompleteName,
	)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), status)
	}
	ctx.JSON(status, resp)
}

func (c *CustomerController) OtpUpdatePhoneMail(ctx *gin.Context) {
	var req OtpUpdatePhoneMailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	resp, status, err := c.CustomerService.OtpUpdatePhoneMail(
		req.ID,
		req.CompleteName,
		req.Email,
		req.PhoneNumber,
		req.CountryCode,
		req.OtpCode,
		req.PhoneChange,
		req.MailChange,
	)
	if err != nil {
		ctx.JSON(status, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(status, resp)
}

func (c *CustomerController) OtpValidatorMail(ctx *gin.Context) {
	var req OtpValidatorMailRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	resp, status, err := c.CustomerService.OtpValidatorMail(
		req.Email,
		req.OtpCode,
	)

	if err != nil {
		ctx.JSON(status, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(status, resp)
}

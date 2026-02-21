package dtos

type OTPValidatorForgotPasswordDTO struct {
	Email   string `json:"email" binding:"required,email"`
	OTPCode string `json:"otp_code" binding:"required"`
}

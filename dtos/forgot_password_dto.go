package dtos

type ForgotPasswordDTO struct {
	Email    string `json:"email" binding:"required,email"`
	OTPType  string `json:"otp_type" binding:"required"`
	Password string `json:"password" binding:"required"`
}

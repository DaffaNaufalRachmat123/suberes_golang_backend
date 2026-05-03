package dtos

type ChangeEmailDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type OtpValidatorEmailVerificationDTO struct {
	ID      string `json:"id" binding:"required"`
	OTPCode string `json:"otp_code" binding:"required"`
	Email   string `json:"email" binding:"required,email"`
}

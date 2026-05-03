package dtos

type OtpValidatorChangePhoneDTO struct {
	ID          string `json:"id" binding:"required"`
	OTPCode     string `json:"otp_code" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	CountryCode string `json:"country_code" binding:"required"`
}

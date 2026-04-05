package repositories

import (
	"suberes_golang/models"

	"gorm.io/gorm"
)

type PinRepository struct {
	DB *gorm.DB
}

type PinStatusResult struct {
	PayPin                          bool   `json:"pay_pin"`
	DisbursementPin                 bool   `json:"disbursement_pin"`
	PayPinStatusText                string `json:"pay_pin_status_text"`
	PayPinSecondStatusText          string `json:"pay_pin_second_status_text"`
	DisbursementPinStatusText       string `json:"disbursement_pin_status_text"`
	DisbursementPinSecondStatusText string `json:"disbursement_pin_second_status_text"`
	PublicKeyPayPin                 string `json:"public_key_pay_pin"`
	PublicKeyDisbursementPin        string `json:"public_key_disbursement_pin"`
}

type PinPublicKeys struct {
	PublicKeyPayPin          string `json:"public_key_pay_pin"`
	PublicKeyDisbursementPin string `json:"public_key_disbursement_pin"`
}

// FindCustomerWithPins loads all pin-related fields for a customer (internal use only).
func (r *PinRepository) FindCustomerWithPins(userID string) (*models.User, error) {
	var user models.User
	err := r.DB.Table("users").
		Select("id, pay_pin, disbursement_pin, private_key_pay_pin, private_key_disbursement_pin, public_key_pay_pin, public_key_disbursement_pin, email").
		Where("id = ? AND user_type = ?", userID, "customer").
		Take(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindPublicKeys returns only RSA public keys for the customer.
func (r *PinRepository) FindPublicKeys(userID string) (*PinPublicKeys, error) {
	var result PinPublicKeys
	err := r.DB.Table("users").
		Select("public_key_pay_pin, public_key_disbursement_pin").
		Where("id = ? AND user_type = ?", userID, "customer").
		Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetPinStatus returns pin status with descriptive text and public keys.
func (r *PinRepository) GetPinStatus(userID string) (*PinStatusResult, error) {
	var result PinStatusResult
	err := r.DB.Table("users").
		Select(`
			CASE WHEN pay_pin IS NULL OR pay_pin = '' THEN false ELSE true END AS pay_pin,
			CASE WHEN disbursement_pin IS NULL OR disbursement_pin = '' THEN false ELSE true END AS disbursement_pin,
			CASE WHEN pay_pin IS NULL OR pay_pin = '' THEN 'PIN Pembayaran Belum Diatur' ELSE 'Pin Pembayaran Sudah Diatur' END AS pay_pin_status_text,
			CASE WHEN pay_pin IS NULL OR pay_pin = '' THEN 'Klik disini untuk atur PIN Pembayaran' ELSE 'Harap klik disini untuk atur PIN Pembayaran' END AS pay_pin_second_status_text,
			CASE WHEN disbursement_pin IS NULL OR disbursement_pin = '' THEN 'PIN Pencairan Belum Diatur' ELSE 'Pin Pencairan Sudah Diatur' END AS disbursement_pin_status_text,
			CASE WHEN disbursement_pin IS NULL OR disbursement_pin = '' THEN 'Klik disini untuk atur PIN Pencairan' ELSE 'Harap klik disini untuk atur PIN Pencairan' END AS disbursement_pin_second_status_text,
			public_key_pay_pin,
			public_key_disbursement_pin
		`).
		Where("id = ? AND user_type = ?", userID, "customer").
		Take(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdatePin updates pay_pin or disbursement_pin for a customer.
func (r *PinRepository) UpdatePin(tx *gorm.DB, userID, field, value string) error {
	return tx.Table("users").
		Where("id = ? AND user_type = ?", userID, "customer").
		Update(field, value).Error
}

// FindOTPForPinChange finds an active OTP record for pin change.
func (r *PinRepository) FindOTPForPinChange(otpID int) (*models.UserOTP, error) {
	var otp models.UserOTP
	err := r.DB.Table("users_otps").
		Where("id = ? AND otp_type = ?", otpID, "change_pin").
		Take(&otp).Error
	if err != nil {
		return nil, err
	}
	return &otp, nil
}

// UpdateOTPCode replaces the otp_code in a record (used in otp_validate).
func (r *PinRepository) UpdateOTPCode(tx *gorm.DB, otpID int, newCode string) error {
	return tx.Table("users_otps").
		Where("id = ? AND otp_type = ?", otpID, "change_pin").
		Update("otp_code", newCode).Error
}

// DestroyOTP deletes an OTP record by id.
func (r *PinRepository) DestroyOTP(tx *gorm.DB, otpID int) error {
	return tx.Table("users_otps").
		Where("id = ? AND otp_type = ?", otpID, "change_pin").
		Delete(nil).Error
}

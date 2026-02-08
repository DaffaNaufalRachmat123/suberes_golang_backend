package repositories

import (
	"suberes_golang/helpers"
	"suberes_golang/models"
	"time"

	"gorm.io/gorm"
)

type UserOtpRepository struct {
	DB *gorm.DB
}

type UserOtp struct {
	OtpCode string
}

func (r *UserOtpRepository) DeleteByUserId(
	tx *gorm.DB,
	userId string,
	payload ...map[string]interface{},
) error {

	query := tx.
		Where("users_id = ?", userId)

	if len(payload) > 0 && payload[0] != nil {
		query = query.Where(payload[0])
	}

	return query.Delete(&models.UserOTP{}).Error
}

func (r *UserOtpRepository) Create(data *models.UserOTP, tx *gorm.DB) error {
	return tx.Create(data).Error
}

func (r *UserOtpRepository) GetValidOtp(
	userId string,
	payload ...map[string]interface{},
) (*UserOtp, error) {

	var otp UserOtp

	validDuration := helpers.GetOtpDuration()
	thresholdTime := time.Now().Add(-validDuration)

	query := r.DB.Table("users_otps").
		Select("otp_code").
		Where("users_id = ?", userId).
		Where("session_time > ?", thresholdTime)

	// kalau payload dipass
	if len(payload) > 0 && payload[0] != nil {
		query = query.Where(payload[0])
	}

	err := query.
		Order("session_time DESC").
		Take(&otp).Error

	if err != nil {
		return nil, err
	}

	return &otp, nil
}

func (r *UserOtpRepository) DestroyUserOtp(userId string, tx *gorm.DB) error {
	return tx.Table("users_otps").Where("id = ?", userId).Delete(nil).Error
}

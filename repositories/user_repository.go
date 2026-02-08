package repositories

import (
	"errors"
	"suberes_golang/models"

	"gorm.io/gorm"
)

type UserRepository struct {
	DB *gorm.DB
}

func (r *UserRepository) FindCustomerById(userId string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ? AND user_type = ?", userId, "customer").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) CheckAvailability(payload map[string]interface{}) (*models.User, error) {
	var user models.User
	err := r.DB.Where(payload).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) UpdateFirebaseToken(tx *gorm.DB, userId, firebaseToken string) error {
	return tx.Table("users").Where("id = ? AND user_type = ?", userId, "customer").Update("firebase_token", firebaseToken).Error
}

func (r *UserRepository) FindCustomerByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("email = ? AND user_type = ?", email, "customer").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) FindPhoneByCustomerEmail(
	phone, email string,
) (*models.User, error) {
	var user models.User
	err := r.DB.Where(
		"phone_number = ? AND email = ? AND user_type = ?", phone, email, "customer",
	).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) CreateUser(tx *gorm.DB, user *models.User) error {
	return tx.Create(user).Error
}
func (r *UserRepository) CustomerProfile(userId string) (*models.User, error) {
	var user models.User
	err := r.DB.Table("users").Select(`
		id,
		complete_name,
		phone_number,
		email,
		country_code
	`).Where("id = ? AND user_type = ?", userId, "customer").Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
	}
	return &user, nil
}
func (r *UserRepository) FindCustomerOtpByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Table("users").Select(`
		id,
		complete_name,
		country_code,
		email,
		phone_number,
		user_type,
		user_rating,
		is_logged_in,
		CASE WHEN pay_pin IS NULL OR pay_pin = '' THEN false ELSE true END AS pay_pin,
			CASE WHEN disbursement_pin IS NULL OR disbursement_pin = '' THEN false ELSE true END AS disbursement_pin
	`).Where("email = ? AND user_type = ?", email, "customer").Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
func (r *UserRepository) UpdateSharedKeys(
	tx *gorm.DB,
	userId string,
	shared_prime, shared_base, shared_secret int64,
) error {
	return tx.Table("users").Where("id = ?", userId).Updates(map[string]interface{}{
		"shared_prime":  shared_prime,
		"shared_base":   shared_base,
		"shared_secret": shared_secret,
	}).Error
}

func (r *UserRepository) UpdateUserData(
	tx *gorm.DB,
	userId string,
	payloadUpdate map[string]interface{},
) (*models.User, error) {

	// 1. Update
	if err := tx.Table("users").
		Where("id = ?", userId).
		Updates(payloadUpdate).Error; err != nil {
		return nil, err
	}

	// 2. Ambil ulang user
	var user models.User
	if err := tx.Table("users").
		Where("id = ?", userId).
		Take(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) SetLoggedIn(tx *gorm.DB, userId string) error {
	return tx.Table("users").Where("id = ? AND user_type = ?", userId, "customer").Update("is_logged_in", "1").Error
}

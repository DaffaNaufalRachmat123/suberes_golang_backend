package repositories

import (
	"gorm.io/gorm"
	"suberes_golang/models"
	"time"
)

type AdminRepository struct {
	DB *gorm.DB
}

func (r *AdminRepository) CountYesterdayMitra(yesterdayStart time.Time, yesterdayEnd time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.User{}).
		Where("createdAt > ? AND createdAt < ? AND is_mitra_activated = '1'", yesterdayStart, yesterdayEnd).
		Count(&count).Error
	return count, err
}

func (r *AdminRepository) CountNewMitra(todayStart time.Time, todayEnd time.Time) (int64, error) {
	var count int64
	err := r.DB.Model(&models.User{}).
		Where("createdAt > ? AND createdAt < ? AND is_mitra_activated = '1'", todayStart, todayEnd).
		Count(&count).Error
	return count, err
}

func (r *AdminRepository) GetAdmins(page int, limit int, userID string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := r.DB.Model(&models.User{})
	db = db.Where("user_type IN (?)", []string{"superadmin", "admin"})
	db = db.Where("id != ?", userID)

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = db.Offset(offset).Limit(limit).Order("complete_name ASC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *AdminRepository) FindAdminByEmail(email string, userType string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("email = ? AND user_type = ?", email, userType).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepository) CreateAdmin(tx *gorm.DB, user *models.User) error {
	return tx.Create(user).Error
}

func (r *AdminRepository) FindAdminByID(id string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ? AND user_type IN (?)", id, []string{"admin", "superadmin"}).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepository) UpdateAdmin(tx *gorm.DB, user *models.User) error {
	return tx.Save(user).Error
}

func (r *AdminRepository) DeleteAdmin(tx *gorm.DB, id string, userType string) error {
	return tx.Where("id = ? AND user_type = ?", id, userType).Delete(&models.User{}).Error
}

func (r *AdminRepository) FindUserForLogin(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Select("id", "complete_name", "email", "country_code", "password", "phone_number", "is_logged_in", "user_type", "user_gender", "address", "domisili_address", "user_profile_image").
		Where("email = ? AND user_type IN (?)", email, []string{"admin", "superadmin"}).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AdminRepository) UpdateUser(tx *gorm.DB, user *models.User, data map[string]interface{}) error {
	return tx.Model(user).Updates(data).Error
}

func (r *AdminRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

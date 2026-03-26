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

func (r *UserRepository) FindOnlineAdmins() ([]string, error) {

	var socketIDs []string

	err := r.DB.
		Model(&models.User{}).
		Select("socket_id").
		Where("user_type IN ?", []string{"admin", "superadmin"}).
		Where("is_logged_in = ?", "1").
		Pluck("socket_id", &socketIDs).Error

	if err != nil {
		return nil, err
	}

	return socketIDs, nil
}

func (r *UserRepository) FindMitraById(userId string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ? AND user_type = ?", userId, "mitra").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAdminById(userId string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("id = ? AND user_type = ? OR user_type = ?", userId, "superadmin", "admin").First(&user).Error
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
func (r *UserRepository) GetOnlineAdminTokens(
	tx *gorm.DB,
) ([]string, error) {

	var tokens []string

	if err := tx.
		Table("users").
		Where("user_type = ? AND is_logged_in = ?", "admin", "1").
		Where("firebase_token IS NOT NULL AND firebase_token <> ''").
		Pluck("firebase_token", &tokens).Error; err != nil {
		return nil, err
	}

	return tokens, nil
}
func (r *UserRepository) FindOneDynamicMap(
	selectFields []string,
	conditions map[string]interface{},
) (map[string]interface{}, error) {

	result := make(map[string]interface{})

	query := r.DB.Table("users")

	if len(selectFields) > 0 {
		query = query.Select(selectFields)
	}

	if len(conditions) > 0 {
		query = query.Where(conditions)
	}

	if err := query.Take(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}
func (r *UserRepository) FindMitraPagination(
	page int,
	limit int,
	search string,
) ([]models.User, int64, error) {

	var users []models.User
	var total int64

	offset := (page - 1) * limit

	query := r.DB.Model(&models.User{}).
		Where("user_type = ?", "mitra").
		Where("is_mitra_invited = ?", "1").
		Where("is_mitra_accepted = ?", "1").
		Where("is_mitra_activated IN ?", []string{"0", "1"}).
		Where("is_suspended IN ?", []string{"0", "1"})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			r.DB.Where("complete_name LIKE ?", searchPattern).
				Or("email LIKE ?", searchPattern).
				Or("phone_number LIKE ?", searchPattern),
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
func (r *UserRepository) FindMitraWithFilter(
	page int,
	limit int,
	search string,
	filters map[string]string,
) ([]models.User, int64, error) {

	var users []models.User
	var total int64

	offset := (page - 1) * limit

	query := r.DB.Model(&models.User{})

	query = query.Where("is_mitra_activated = ?", "0").
		Where("user_type = ?", "mitra")

	for key, value := range filters {
		if value != "" {
			query = query.Where(key+" = ?", value)
		}
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			r.DB.Where("complete_name LIKE ?", searchPattern).
				Or("email LIKE ?", searchPattern).
				Or("phone_number LIKE ?", searchPattern),
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
func (r *UserRepository) FindGenderMitraExGoLife(id string) (*models.User, error) {
	var user models.User

	err := r.DB.
		Select(`
			users.*,
			CASE user_gender 
				WHEN 'male' THEN 'Pria' 
				ELSE 'Wanita' 
			END as user_gender,
			CASE is_ex_golife 
				WHEN '1' THEN 'Ya' 
				ELSE 'Tidak' 
			END as is_ex_golife
		`).
		Where("id = ? AND user_type = ?", id, "mitra").
		First(&user).Error

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

func (r *UserRepository) FindMitraByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.DB.Where("email = ? AND user_type = ?", email, "mitra").First(&user).Error
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

func (r *UserRepository) MitraProfile(userId string) (*models.User, error) {
	var user models.User
	err := r.DB.Table("users").Select(`
		id,
		user_profile_image,
		complete_name,
		email,
		is_mitra_activated,
		user_level,
		color_code_level,
		country_code,
		phone_number,
		user_rating,
		user_gender,
		place_of_birth,
		TIMESTAMPDIFF(YEAR , date_of_birth , CURDATE()) as age
	`).Where("id = ? AND user_type = ?", userId, "mitra").Take(&user).Error
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

func (r *UserRepository) AddCustomerBalance(tx *gorm.DB, customerID string, grossAmount float64) error {
	result := tx.Model(&models.User{}).Where("id = ? AND user_type = ?", customerID, "customer").Update("account_balance", gorm.Expr("account_balance + ?", grossAmount))
	return result.Error
}

func (r *UserRepository) UpdateUserData(
	tx *gorm.DB,
	where map[string]interface{},
	payloadUpdate map[string]interface{},
) (*models.User, error) {

	// 1. Update
	if err := tx.Table("users").
		Where(where).
		Updates(payloadUpdate).Error; err != nil {
		return nil, err
	}

	// 2. Ambil ulang user
	var user models.User
	if err := tx.Table("users").
		Where(where).
		Take(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}
func (r *UserRepository) SetLoggedIn(tx *gorm.DB, userId string) error {
	return tx.Table("users").Where("id = ? AND user_type = ?", userId, "customer").Update("is_logged_in", "1").Error
}

func (r *UserRepository) SetMitraLoggedIn(tx *gorm.DB, userId string) error {
	return tx.Table("users").Where("id = ? AND user_type = ?", userId, "mitra").Update("is_logged_in", "1").Error
}
func (r *UserRepository) DeleteByConditions(
	tx *gorm.DB,
	conditions map[string]interface{},
) error {

	query := tx.Table("users")

	for key, value := range conditions {
		query = query.Where(key+" = ?", value)
	}

	return query.Delete(nil).Error
}

func (r *UserRepository) FindUserForTransaction(tx *gorm.DB, transaction *models.Transaction) (*models.User, error) {
	var user models.User
	userID := transaction.CustomerID
	userType := "customer"

	if userID == "" || userID == "NULL" {
		userID = transaction.MitraID
		userType = "mitra"
	}

	err := tx.Where("id = ? AND user_type = ?", userID, userType).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUserBalance(tx *gorm.DB, userID string, amount int64) error {
	return tx.Model(&models.User{}).Where("id = ?", userID).Update("account_balance", gorm.Expr("account_balance + ?", amount)).Error
}

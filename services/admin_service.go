package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminService struct {
	OrderRepo *repositories.OrderTransactionRepository
	AdminRepo *repositories.AdminRepository
	DB        *gorm.DB
}

func (s *AdminService) GetDashboard() (*dtos.DashboardPayload, error) {

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := todayStart.Add(24*time.Hour - time.Second)

	yesterdayStart := todayStart.Add(-24 * time.Hour)
	yesterdayEnd := todayStart.Add(-time.Second)

	yesterdayOrders, _ := s.OrderRepo.CountOrders(yesterdayStart, yesterdayEnd)
	todayOrders, _ := s.OrderRepo.CountFinishedOrders(todayStart, todayEnd)

	yesterdayRevenue, _ := s.OrderRepo.SumRevenue(yesterdayStart, yesterdayEnd)
	todayRevenue, _ := s.OrderRepo.SumRevenue(todayStart, todayEnd)

	yesterdayMitra, _ := s.AdminRepo.CountYesterdayMitra(yesterdayStart, yesterdayEnd)
	todayMitra, _ := s.AdminRepo.CountNewMitra(todayStart, todayEnd)

	totalOrdersByMonth, _ := s.OrderRepo.TotalOrdersByMonth()
	overviewMonthRevenue, _ := s.OrderRepo.OverviewMonthRevenue()
	overviewWeekRevenue, _ := s.OrderRepo.OverviewWeekRevenue()
	frequentlyUsedService, _ := s.OrderRepo.FrequentlyUsedService()
	mitraOrderData, _ := s.OrderRepo.MitraOrderToday(todayStart, todayEnd)

	payload := &dtos.DashboardPayload{}

	payload.TodayOrderData.TodayCount = todayOrders
	payload.TodayOrderData.Percentage = calculatePercentage(yesterdayOrders, todayOrders)

	payload.TodayMitraData.TodayCount = todayMitra
	payload.TodayMitraData.Percentage = calculatePercentage(yesterdayMitra, todayMitra)

	payload.TodayTransactionData.TodayCount = todayRevenue
	payload.TodayTransactionData.Percentage = calculatePercentage(yesterdayRevenue, todayRevenue)

	payload.TotalOrdersByMonth = totalOrdersByMonth
	payload.OverviewMonth = overviewMonthRevenue
	payload.OverviewWeek = overviewWeekRevenue
	payload.FusService = frequentlyUsedService
	payload.MitraOrderToday = mitraOrderData

	return payload, nil
}
func calculatePercentage(oldVal, newVal int64) float64 {
	if oldVal == 0 {
		if newVal == 0 {
			return 0
		}
		return 100
	}
	return (float64(newVal-oldVal) / float64(oldVal)) * 100
}

func (s *AdminService) IndexAdmin(page int, limit int, userID string) ([]models.User, int64, error) {
	return s.AdminRepo.GetAdmins(page, limit, userID)
}

func (s *AdminService) CreateAdmin(req *dtos.CreateAdminRequest, fileHeader string) (*models.User, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	existingUser, _ := s.AdminRepo.FindAdminByEmail(req.Email, req.UserType)
	if existingUser != nil {
		return nil, errors.New("email exist")
	}

	plainPass, err := generatePassword(10)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPass), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newUser := &models.User{
		ID:              uuid.New().String(),
		CompleteName:    req.CompleteName,
		Email:           req.Email,
		PhoneNumber:     req.PhoneNumber,
		CountryCode:     req.CountryCode,
		UserType:        req.UserType,
		UserGender:      req.UserGender,
		Address:         req.Alamat,
		DomisiliAddress: req.DomisiliAddress,
		IsActive:        "yes",
		UserLevel:       "no level",
		ColorCodeLevel:  "#CECECE",
		KTPImage:        fileHeader,
		Password:        string(hashedPassword),
	}

	err = s.AdminRepo.CreateAdmin(tx, newUser)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if len(req.UserType) > 0 {
		userType := strings.ToUpper(req.UserType[:1]) + req.UserType[1:]
		helpers.SendAcceptedAdminAccount(os.Getenv("SUPPORT_EMAIL"), req.Email, userType, req.Email, plainPass)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	return newUser, nil
}

func (s *AdminService) UpdateAdminStatus(adminID string, req *dtos.UpdateAdminStatusRequest) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	admin, err := s.AdminRepo.FindAdminByID(adminID)
	if err != nil {
		tx.Rollback()
		return errors.New("admin not found")
	}

	updateData := map[string]interface{}{
		"is_active": req.IsActive,
	}

	if req.IsActive == "yes" {
		plainPass, err := generatePassword(10)
		if err != nil {
			tx.Rollback()
			return err
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPass), bcrypt.DefaultCost)
		if err != nil {
			tx.Rollback()
			return err
		}
		updateData["password"] = string(hashedPassword)
		updateData["activate_reason"] = req.Reason
	} else {
		updateData["is_logged_in"] = "0"
		updateData["socket_id"] = ""
		updateData["nonactivate_reason"] = req.Reason
		updateData["password"] = ""
	}

	if err := s.AdminRepo.UpdateUser(tx, admin, updateData); err != nil {
		tx.Rollback()
		return err
	}

	// TODO: send email
	// if(data.is_active === 'yes'){
	// 	await sendmail.sendActiveAdminAccount(process.env.SUPPORT_EMAIL , admin_data.email , 'Status Akun' , admin_data.complete_name , admin_data.user_type , admin_data.user_type.charAt(0).toUpperCase() + admin_data.user_type.slice(1) , admin_data.email , password_plain , payUpdate.activate_reason)
	// } else if(data.is_active === 'no'){
	// 	await sendmail.sendNonactiveAdminAccount(process.env.SUPPORT_EMAIL , admin_data.email , 'Status Akun' , admin_data.complete_name , admin_data.user_type , admin_data.user_type.charAt(0).toUpperCase() + admin_data.user_type.slice(1) , admin_data.email , payUpdate.nonactivate_reason)
	// }
	tx.Commit()
	return nil
}

func (s *AdminService) RemoveAdmin(adminID string, superAdminID string, req *dtos.RemoveAdminRequest) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	superAdmin, err := s.AdminRepo.FindAdminByID(superAdminID)
	if err != nil {
		tx.Rollback()
		return errors.New("superadmin data not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(superAdmin.Password), []byte(req.Password)); err != nil {
		tx.Rollback()
		return errors.New("unauthorized")
	}

	admin, err := s.AdminRepo.FindAdminByID(adminID)
	if err != nil {
		tx.Rollback()
		return errors.New("admin data not found")
	}

	if err := s.AdminRepo.DeleteAdmin(tx, adminID, admin.UserType); err != nil {
		tx.Rollback()
		return err
	}

	// TODO: send email
	// await sendmail.sendRemoveAdminAccount(process.env.SUPPORT_EMAIL , admin_data.email , 'Penghapusan Akun' , admin_data.complete_name , admin_data.user_type.charAt(0) + admin_data.user_type.slice(1) , admin_data.email , data.reason)

	tx.Commit()
	return nil
}

func (s *AdminService) RefreshToken(userId string) (string, error) {
	user, err := s.AdminRepo.FindAdminByID(userId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("user not found")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"id":               user.ID,
		"complete_name":    user.CompleteName,
		"email":            user.Email,
		"phone_number":     user.PhoneNumber,
		"country_code":     user.CountryCode,
		"user_type":        user.UserType,
		"user_gender":      user.UserGender,
		"address":          user.Address,
		"domisili_address": user.DomisiliAddress,
		"issued_at":        now.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (s *AdminService) Login(req *dtos.LoginAdminRequest) (string, *models.User, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	user, err := s.AdminRepo.FindUserForLogin(req.Email)
	if err != nil {
		tx.Rollback()
		return "", nil, errors.New("akun tidak ditemukan")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		tx.Rollback()
		return "", nil, errors.New("password salah")
	}

	if err := s.AdminRepo.UpdateUser(tx, user, map[string]interface{}{"is_logged_in": "1"}); err != nil {
		tx.Rollback()
		return "", nil, err
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"id":               user.ID,
		"complete_name":    user.CompleteName,
		"email":            user.Email,
		"phone_number":     user.PhoneNumber,
		"country_code":     user.CountryCode,
		"user_type":        user.UserType,
		"user_gender":      user.UserGender,
		"address":          user.Address,
		"domisili_address": user.DomisiliAddress,
		"issued_at":        now.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		tx.Rollback()
		return "", nil, err
	}

	tx.Commit()
	return tokenString, user, nil
}

func (s *AdminService) Logout(userID string) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	user, err := s.AdminRepo.FindAdminByID(userID)
	if err != nil {
		tx.Rollback()
		return errors.New("admin data not found")
	}

	if err := s.AdminRepo.UpdateUser(tx, user, map[string]interface{}{"is_logged_in": "0", "socket_id": "", "firebase_token": ""}); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (s *AdminService) UpdateFirebaseToken(userID string, token string) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	user, err := s.AdminRepo.FindAdminByID(userID)
	if err != nil {
		tx.Rollback()
		return errors.New("user not found")
	}

	if err := s.AdminRepo.UpdateUser(tx, user, map[string]interface{}{"firebase_token": token}); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func generatePassword(length int) (string, error) {
	const wishlist = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz~!@-_+=#$"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = wishlist[b%byte(len(wishlist))]
	}
	return string(bytes), nil
}
func (s *AdminService) GetUserByID(id string) (*models.User, error) {
	var user models.User
	err := s.DB.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	return &user, nil
}

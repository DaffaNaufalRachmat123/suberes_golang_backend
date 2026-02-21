package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CustomerService struct {
	UserRepo    *repositories.UserRepository
	UserOTPRepo *repositories.UserOtpRepository
	DB          *gorm.DB
}

func (s *CustomerService) OtpUpdatePhoneMail(
	ID string,
	CompleteName string,
	Email string,
	PhoneNumber string,
	CountryCode string,
	OtpCode string,
	PhoneChange bool,
	MailChange bool,
) (map[string]interface{}, int, error) {
	user, _ := s.UserRepo.FindCustomerById(ID)
	if user == nil {
		return nil, 404, errors.New("Customer not found")
	}
	payloadCheck := map[string]interface{}{
		"user_type": "customer",
	}
	if PhoneChange {
		payloadCheck["phone_number"] = PhoneNumber
	}
	if MailChange {
		payloadCheck["email"] = Email
	}
	existUser, _ := s.UserRepo.CheckAvailability(payloadCheck)
	if existUser != nil {
		return nil, 409, errors.New("Email or phone number already registered")
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	otpData, _ := s.UserOTPRepo.GetValidOtp(ID, map[string]interface{}{
		"otp_type": "change_data",
	})
	if otpData == nil {
		tx.Rollback()
		return nil, 400, errors.New("Your OTP Number is out of time")
	}
	if bcrypt.CompareHashAndPassword(
		[]byte(otpData.OtpCode),
		[]byte(OtpCode),
	) != nil {
		return nil, 401, errors.New("OTP code is wrong")
	}
	payloadUpdate := make(map[string]interface{})
	if PhoneChange {
		payloadUpdate["phone_number"] = PhoneNumber
		payloadUpdate["country_code"] = CountryCode
	}
	if MailChange {
		payloadUpdate["email"] = Email
	}
	user, err := s.UserRepo.UpdateUserData(tx, map[string]interface{}{
		"id": ID,
	}, payloadUpdate)
	if err != nil {
		return nil, 500, errors.New("Failed to update user data")
	}
	claims := jwt.MapClaims{
		"id":            user.ID,
		"complete_name": user.CompleteName,
		"country_code":  user.CountryCode,
		"email":         user.Email,
		"phone_number":  user.PhoneNumber,
		"user_type":     user.UserType,
		"user_rating":   user.UserRating,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err := s.UserOTPRepo.DeleteByUserId(tx, ID, map[string]interface{}{
		"otp_type": "change_data",
	}); err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	tx.Commit()

	return map[string]interface{}{
		"server_message": "change data succeed",
		"status":         "OK",
		"token":          tokenString,
		"data":           user,
	}, 200, nil
}

func (s *CustomerService) ChangePhoneMail(
	ID string,
	phoneChange bool,
	mailChange bool,
	phoneNumber string,
	email string,
) (map[string]interface{}, int, error) {
	userData, err := s.UserRepo.FindCustomerById(ID)
	if err != nil {
		return nil, 404, errors.New("Customer not found")
	}
	payloadCheck := map[string]interface{}{
		"user_type": "customer",
	}
	if phoneChange {
		payloadCheck["phone_number"] = phoneNumber
	}
	if mailChange {
		payloadCheck["email"] = email
	}
	existingUser, _ := s.UserRepo.CheckAvailability(payloadCheck)
	if existingUser != nil {
		return nil, 409, errors.New("Email or phone already registered by another account")
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, 500, tx.Error
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	otpCode := fmt.Sprintf("%06d", rng.Intn(900000)+100000)
	hashedOtp, err := bcrypt.GenerateFromPassword([]byte(otpCode), 12)
	if err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	if err := s.UserOTPRepo.DeleteByUserId(tx, ID); err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	newOtp := models.UserOTP{
		UsersID:     userData.ID,
		OTPCode:     string(hashedOtp),
		OTPType:     "change_data",
		SessionTime: time.Now(), // Sequelize.fn('now')
	}

	if err := s.UserOTPRepo.Create(&newOtp, tx); err != nil {
		tx.Rollback()
		return nil, 500, err
	}

	// 8. Kirim Email
	// Menggunakan helper SendOtpCodeMail yang sudah dibuat sebelumnya
	go helpers.SendOtpCodeMail(os.Getenv("SUPPORT_EMAIL"), email, "Kode OTP Suberes", otpCode)

	fmt.Printf("OTP Code : %s\n", otpCode)

	// 9. Commit Transaksi
	if err := tx.Commit().Error; err != nil {
		return nil, 500, err
	}

	// 10. Return Success Response
	return map[string]interface{}{
		"server_message": "otp number sent",
		"otp_timeout":    os.Getenv("OTP_TIMEOUT"),
		"status":         "success",
	}, 200, nil
}

func (s *CustomerService) GetCustomerProfile(userId string) (interface{}, int, error) {
	user, err := s.UserRepo.CustomerProfile(userId)
	if err != nil {
		return nil, 500, err
	}
	if user == nil {
		return nil, 404, errors.New("Customer not found")
	}
	return &user, 200, nil
}

func (s *CustomerService) Register(
	completeName string,
	email string,
	phoneNumber string,
	countryCode string,
	userType string,
) error {
	tx := s.DB.Begin()
	if tx.Error != nil {
		return errors.New("Failed to start the transaction")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	_, err := s.UserRepo.FindPhoneByCustomerEmail(phoneNumber, email)
	if err == nil {
		tx.Rollback()
		return errors.New("Phone number or email already registered")
	}
	CountryCode := countryCode
	if !strings.HasPrefix(countryCode, "+") {
		CountryCode = "+" + countryCode
	}
	pinPrivateKey, pinPublicKey, err := helpers.GenerateRSAKeyPair()
	if err != nil {
		return err
	}

	disbursementPrivateKey, disbursementPublicKey, err := helpers.GenerateRSAKeyPair()
	if err != nil {
		return err
	}
	user := models.User{
		ID:                        uuid.NewString(),
		CompleteName:              completeName,
		Email:                     email,
		PhoneNumber:               phoneNumber,
		CountryCode:               CountryCode,
		UserType:                  "customer",
		UserLevel:                 "no level",
		ColorCodeLevel:            "#CECECE",
		PrivateKeyPayPin:          pinPrivateKey,
		PublicKeyPayPin:           pinPublicKey,
		PrivateKeyDisbursementPin: disbursementPrivateKey,
		PublicKeyDisbursementPin:  disbursementPublicKey,
	}
	if err := s.UserRepo.CreateUser(tx, &user); err != nil {
		tx.Rollback()
		return err
	}

	// generate OTP
	otpCode := rand.Intn(900000) + 100000

	hashedOtp, _ := bcrypt.GenerateFromPassword(
		[]byte(strconv.Itoa(otpCode)),
		12,
	)

	// hapus OTP lama
	_ = s.UserOTPRepo.DeleteByUserId(tx, user.ID)

	otp := models.UserOTP{
		UsersID:     user.ID,
		OTPCode:     string(hashedOtp),
		OTPType:     "email_verification_code",
		SessionTime: time.Now(),
	}

	if err := s.UserOTPRepo.Create(&otp, tx); err != nil {
		tx.Rollback()
		return err
	}

	// kirim OTP async
	go helpers.SendOtpCodeMail(
		os.Getenv("SUPPORT_EMAIL"),
		email,
		"Kode OTP Suberes",
		strconv.Itoa(otpCode),
	)

	tx.Commit()
	return nil
}

func (s *CustomerService) LoginByEmail(email string) (string, error) {
	tx := s.DB.Begin()

	user, err := s.UserRepo.FindCustomerByEmail(email)
	if err != nil {
		tx.Rollback()
		return "", errors.New("CUSTOMER_NOT_FOUND")
	}

	if user.IsLoggedIn == "1" {
		tx.Rollback()
		return "", errors.New("CUSTOMER_ALREADY_LOGGED_IN")
	}

	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000

	fmt.Println("OTP Code", otp)

	hash, _ := bcrypt.GenerateFromPassword([]byte(strconv.Itoa(otp)), 12)

	s.UserOTPRepo.DeleteByUserId(tx, user.ID)

	err = s.UserOTPRepo.Create(&models.UserOTP{
		UsersID:     user.ID,
		OTPCode:     string(hash),
		OTPType:     "email_verification_code",
		SessionTime: time.Now(),
	}, tx)

	if err != nil {
		tx.Rollback()
		return "", err
	}

	go helpers.SendOtpCodeMail(
		os.Getenv("SUPPORT_EMAIL"),
		email,
		"Kode OTP Suberes",
		strconv.Itoa(otp),
	)

	tx.Commit()

	return strconv.Itoa(otp), nil
}

func (s *CustomerService) UpdateFirebaseTokenCustomer(userId, firebaseToken string) (map[string]interface{}, int, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	user, err := s.UserRepo.FindCustomerById(userId)
	if err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	if user == nil {
		tx.Rollback()
		return nil, 404, errors.New("Customer not found")
	}
	if err := s.UserRepo.UpdateFirebaseToken(tx, userId, firebaseToken); err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, 500, err
	}
	return map[string]interface{}{
		"server_message": "Firebase token updated",
		"status":         "OK",
	}, 200, nil
}

func (s *CustomerService) OtpValidatorMail(email, otpCode string) (map[string]interface{}, int, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	user, err := s.UserRepo.FindCustomerOtpByEmail(email)
	if err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	if user == nil {
		tx.Rollback()
		return nil, 404, errors.New("Email is not registered")
	}
	if user.IsLoggedIn == "1" {
		tx.Rollback()
		return nil, 401, errors.New("User already logged on other device, please logout first")
	}
	otp, err := s.UserOTPRepo.GetValidOtp(user.ID)
	if err != nil {
		tx.Rollback()
		return nil, 400, errors.New("Your OTP number is out of time")
	}
	if bcrypt.CompareHashAndPassword(
		[]byte(otp.OtpCode),
		[]byte(otpCode),
	) != nil {
		tx.Rollback()
		return nil, 401, errors.New("OTP code is wrong")
	}
	sharedPrime := helpers.GetRandomPrimeNumber()
	sharedBase := helpers.FindPrimitiveRoot(sharedPrime)
	sharedSecret := helpers.GetRandomPrimeNumber()
	if err := s.UserRepo.UpdateSharedKeys(
		tx,
		user.ID,
		sharedPrime,
		sharedBase,
		sharedSecret,
	); err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	claims := jwt.MapClaims{
		"id":            user.ID,
		"complete_name": user.CompleteName,
		"country_code":  user.CountryCode,
		"email":         user.Email,
		"phone_number":  user.PhoneNumber,
		"user_type":     user.UserType,
		"user_rating":   user.UserRating,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err := s.UserOTPRepo.DeleteByUserId(tx, user.ID); err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	if err := s.UserRepo.SetLoggedIn(tx, user.ID); err != nil {
		tx.Rollback()
		return nil, 500, err
	}

	tx.Commit()

	return map[string]interface{}{
		"server_message": "login successfully",
		"status":         "success",
		"token":          "Bearer " + tokenString,
		"data":           user,
		"shared_prime":   sharedPrime,
		"shared_secret":  sharedSecret,
	}, 200, nil
}

func (s *CustomerService) UpdateUserProfile(
	ID string,
	CompleteName string,
) (map[string]interface{}, int, error) {
	user, err := s.UserRepo.FindCustomerById(ID)
	if err != nil {
		return nil, 500, err
	}
	if user == nil {
		return nil, 404, errors.New("Customer not found")
	}
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			return
		}
	}()
	updatedUser, err := s.UserRepo.UpdateUserData(tx, map[string]interface{}{
		"id": ID,
	}, map[string]interface{}{
		"complete_name": CompleteName,
	})
	return map[string]interface{}{
		"server_message": "Profile updated",
		"status":         "OK",
		"data":           updatedUser,
	}, 200, nil
}

func (s *CustomerService) Logout(ID string) (map[string]interface{}, int, error) {
	user, err := s.UserRepo.FindCustomerByEmail(ID)
	if err != nil {
		return nil, 500, err
	}
	if user == nil {
		return nil, 404, errors.New("Customer not found")
	}
	tx := s.DB.Begin()
	_, err = s.UserRepo.UpdateUserData(tx, map[string]interface{}{
		"id": ID,
	}, map[string]interface{}{
		"firebase_token": "",
		"is_logged_in":   "0",
	})
	if err != nil {
		return nil, 500, errors.New("Failed to update user data")
	}
	if err != nil {
		tx.Rollback()
		return nil, 500, err
	}
	tx.Commit()
	return map[string]interface{}{
		"server_message": "Logout succeed",
		"status":         "OK",
	}, 200, nil
}

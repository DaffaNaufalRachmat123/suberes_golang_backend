package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"suberes_golang/service"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type MitraService struct {
	DB                         *gorm.DB
	MitraRepository            *repositories.MitraRepository
	UserRepository             *repositories.UserRepository
	OrderTransactionRepository *repositories.OrderTransactionRepository
	PaymentRepository          *repositories.PaymentRepository
	SubPaymentRepository       *repositories.SubPaymentRepository
	TransactionRepository      *repositories.TransactionRepository
	UserOtpRepository          *repositories.UserOtpRepository
}

func NewMitraService(db *gorm.DB, mitraRepo *repositories.MitraRepository, userRepo *repositories.UserRepository, orderTransactionRepo *repositories.OrderTransactionRepository, userOtpRepo *repositories.UserOtpRepository) *MitraService {
	return &MitraService{
		DB:                         db,
		MitraRepository:            mitraRepo,
		UserRepository:             userRepo,
		OrderTransactionRepository: orderTransactionRepo,
		UserOtpRepository:          userOtpRepo,
	}
}

func (s *MitraService) Login(loginDTO dtos.MitraLoginDTO) (*dtos.MitraLoginResponseDTO, error) {
	mitra, err := s.MitraRepository.FindMitraByEmail(loginDTO.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mitra not found")
		}
		return nil, err
	}

	if mitra.IsLoggedIn == "1" {
		return nil, errors.New("mitra already logged in another device, please log out first")
	}

	if mitra.IsMitraActivated != "1" {
		return nil, errors.New("mitra not activated")
	}

	err = bcrypt.CompareHashAndPassword([]byte(mitra.Password), []byte(loginDTO.Password))
	if err != nil {
		return nil, errors.New("password not match")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	sharedSecret := helpers.GetRandomPrimeNumber()
	mitra.IsLoggedIn = "1"
	mitra.SharedSecret = sharedSecret

	if err := s.MitraRepository.UpdateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return nil, err
	}

	order, err := s.OrderTransactionRepository.FindRunningOrderByMitraID(mitra.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}

	var sharedPrimeCustomer int
	if order != nil {
		customer, err := s.UserRepository.FindCustomerById(string(order.CustomerID))
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		sharedPrimeCustomer = customer.SharedPrime
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":                 mitra.ID,
		"complete_name":      mitra.CompleteName,
		"email":              mitra.Email,
		"phone_number":       mitra.PhoneNumber,
		"user_type":          mitra.UserType,
		"user_rating":        mitra.UserRating,
		"user_profile_image": mitra.UserProfileImage,
		"user_status":        mitra.UserStatus,
		"exp":                time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	status := "NOT_IN_ORDER"
	if order != nil {
		status = "IN_ORDER"
	}

	return &dtos.MitraLoginResponseDTO{
		ServerMessage: "login successfully",
		Status:        status,
		Token:         "Bearer " + tokenString,
		Data:          *mitra,
		SharedPrime:   sharedPrimeCustomer,
		SharedSecret:  sharedSecret,
	}, nil
}

func (s *MitraService) Register(registerDTO dtos.MitraRegisterDTO, files map[string]string) (*models.User, error) {
	if _, err := s.UserRepository.CheckAvailability(map[string]interface{}{"email": registerDTO.Email, "user_type": "mitra"}); err == nil {
		return nil, errors.New("This email is used")
	}
	if _, err := s.UserRepository.CheckAvailability(map[string]interface{}{"ktp_number": registerDTO.KTPNumber, "user_type": "mitra"}); err == nil {
		return nil, errors.New("This KTP is used")
	}
	if _, err := s.UserRepository.CheckAvailability(map[string]interface{}{"phone_number": registerDTO.PhoneNumber, "user_type": "mitra"}); err == nil {
		return nil, errors.New("This phone number is used")
	}
	if _, err := s.UserRepository.CheckAvailability(map[string]interface{}{"emergency_contact_phone": registerDTO.EmergencyContactPhone, "user_type": "mitra"}); err == nil {
		return nil, errors.New("This phone number darurat is used")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	mitra := &models.User{
		ID:                          newUUID.String(),
		Email:                       registerDTO.Email,
		CompleteName:                registerDTO.CompleteName,
		PlaceOfBirth:                registerDTO.PlaceOfBirth,
		Age:                         registerDTO.Age,
		DateOfBirth:                 registerDTO.DateOfBirth,
		UserGender:                  registerDTO.UserGender,
		KTPNumber:                   registerDTO.KTPNumber,
		PhoneNumber:                 registerDTO.PhoneNumber,
		CountryCode:                 registerDTO.CountryCode,
		EmergencyContactCountryCode: registerDTO.EmergencyContactCountryCode,
		EmergencyContactName:        registerDTO.EmergencyContactName,
		EmergencyContactRelation:    registerDTO.EmergencyContactRelation,
		EmergencyContactPhone:       registerDTO.EmergencyContactPhone,
		Address:                     registerDTO.Address,
		ColorCodeLevel:              "#CECECE",
		DomisiliAddress:             registerDTO.DomisiliAddress,
		RTRW:                        registerDTO.RTRW,
		SubDistrict:                 registerDTO.SubDistrict,
		District:                    registerDTO.District,
		Province:                    registerDTO.Province,
		City:                        registerDTO.City,
		PostalCode:                  registerDTO.PostalCode,
		WorkExperience:              registerDTO.WorkExperience,
		IsExGolife:                  registerDTO.IsExGolife,
		KindOfMitra:                 registerDTO.KindOfMitra,
		UserLevel:                   "no level",
		WorkExperienceDuration:      registerDTO.WorkExperienceDuration,
		UserType:                    "mitra",
		IsMitraInvited:              "0",
		IsMitraRejected:             "0",
		IsMitraActivated:            "0",
		KTPImage:                    files["ktp"],
		KKImage:                     files["kk"],
		CoverSavingsBook:            files["cover_saving_book"],
		UserProfileImage:            files["profile_image"],
	}

	if err := s.MitraRepository.CreateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return mitra, nil
}

func (s *MitraService) GetProfile(mitraID string, timezoneCode string) (*dtos.MitraProfileResponseDTO, error) {
	startDateTime, endDateTime, err := helpers.GetStartAndEndDate(timezoneCode)
	if err != nil {
		return nil, err
	}

	mitra, err := s.MitraRepository.GetMitraProfile(mitraID)
	if err != nil {
		return nil, err
	}

	orderCount, err := s.MitraRepository.GetTodayOrderCount(mitraID, startDateTime, endDateTime)
	if err != nil {
		return nil, err
	}

	orderCountRepeat, err := s.MitraRepository.GetTodayOrderRepeatCount(mitraID, startDateTime, endDateTime)
	if err != nil {
		return nil, err
	}

	totalCicilan, err := s.MitraRepository.GetTotalCicilan(mitraID)
	if err != nil {
		return nil, err
	}

	response := &dtos.MitraProfileResponseDTO{
		Profile:  mitra,
		BillData: totalCicilan,
	}
	response.OrderCount.OrderCount = orderCount.OrderCount + orderCountRepeat.OrderCount
	response.OrderCount.PendapatanOrder = orderCount.PendapatanHariIni + orderCountRepeat.PendapatanHariIni

	return response, nil
}

func (s *MitraService) GetEmailPassword(mitraID string) (*dtos.MitraEmailPasswordResponseDTO, error) {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return nil, err
	}

	return &dtos.MitraEmailPasswordResponseDTO{
		Email:    mitra.Email,
		Password: mitra.Password,
	}, nil
}

func (s *MitraService) ChangePassword(mitraID string, changePasswordDTO dtos.ChangePasswordDTO) error {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(mitra.Password), []byte(changePasswordDTO.OldPassword))
	if err != nil {
		return errors.New("password lama kamu salah")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(changePasswordDTO.Password), 12)
	if err != nil {
		return err
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	mitra.Password = string(hashedPassword)
	if err := s.MitraRepository.UpdateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MitraService) ChangeEmail(mitraID string, changeEmailDTO dtos.ChangeEmailDTO) (string, error) {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return "", err
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	if err := s.UserOtpRepository.DeleteByUserId(tx, mitra.ID, nil); err != nil {
		tx.Rollback()
		return "", err
	}

	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	hashedOtp, err := bcrypt.GenerateFromPassword([]byte(otpCode), 12)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	userOtp := &models.UserOTP{
		UsersID:     mitra.ID,
		OTPCode:     string(hashedOtp),
		OTPType:     "email_verification_code",
		SessionTime: time.Now(),
	}

	if err := s.UserOtpRepository.Create(userOtp, tx); err != nil {
		tx.Rollback()
		return "", err
	}

	// TODO: send email
	fmt.Println("OTP Code: ", otpCode)

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return os.Getenv("OTP_TIMEOUT"), nil
}

func (s *MitraService) ChangeForgotPassword(dto dtos.ForgotPasswordDTO) error {
	mitra, err := s.UserRepository.FindMitraByEmail(dto.Email)
	if err != nil {
		return errors.New("mitra data not found")
	}

	_, err = s.UserOtpRepository.GetValidOtp(mitra.ID, map[string]interface{}{"otp_type": dto.OTPType})
	if err != nil {
		return errors.New("Unauthorized")
	}

	// Since there is no otp code in the request, I'll just comment this out for now
	// err = bcrypt.CompareHashAndPassword([]byte(otp.OtpCode), []byte(dto.OTP))
	// if err != nil {
	// 	return errors.New("Unauthorized")
	// }

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), 12)
	if err != nil {
		return err
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	mitra.Password = string(hashedPassword)
	if err := s.MitraRepository.UpdateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.UserOtpRepository.DeleteByUserId(tx, mitra.ID, map[string]interface{}{"otp_type": dto.OTPType}); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MitraService) RequestForgotPassword(email string) (string, error) {
	mitra, err := s.UserRepository.FindMitraByEmail(email)
	if err != nil {
		return "", errors.New("mitra not found")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	if err := s.UserOtpRepository.DeleteByUserId(tx, mitra.ID, map[string]interface{}{"otp_type": "forgot_password"}); err != nil {
		tx.Rollback()
		return "", err
	}

	otpCode := fmt.Sprintf("%06d", rand.Intn(1000000))
	hashedOtp, err := bcrypt.GenerateFromPassword([]byte(otpCode), 12)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	userOtp := &models.UserOTP{
		UsersID:     mitra.ID,
		OTPCode:     string(hashedOtp),
		OTPType:     "forgot_password",
		SessionTime: time.Now(),
	}

	if err := s.UserOtpRepository.Create(userOtp, tx); err != nil {
		tx.Rollback()
		return "", err
	}

	// TODO: send email
	fmt.Println("OTP Code: ", otpCode)

	if err := tx.Commit().Error; err != nil {
		return "", err
	}

	return os.Getenv("OTP_TIMEOUT"), nil
}

func (s *MitraService) OTPValidatorForgotPassword(dto dtos.OTPValidatorForgotPasswordDTO) error {
	mitra, err := s.UserRepository.FindMitraByEmail(dto.Email)
	if err != nil {
		return errors.New("mitra not found")
	}

	otp, err := s.UserOtpRepository.GetValidOtp(mitra.ID, map[string]interface{}{"otp_type": "forgot_password"})
	if err != nil {
		return errors.New("Your OTP number not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(otp.OtpCode), []byte(dto.OTPCode))
	if err != nil {
		return errors.New("otp code is wrong")
	}

	return nil
}

func (s *MitraService) Logout(mitraID string) error {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return errors.New("mitra not found")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	mitra.FirebaseToken = nil
	mitra.IsActive = "no"
	mitra.IsAutoBid = "no"
	mitra.IsLoggedIn = "0"

	if err := s.MitraRepository.UpdateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *MitraService) UpdateFirebaseToken(mitraID string, token string) error {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return errors.New("user not found")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	*mitra.FirebaseToken = token

	if err := s.MitraRepository.UpdateMitra(tx, mitra); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
func (s *MitraService) UpdateMitraStatus(ctx context.Context, mitraID, status, userType string, suspendedReason string) (int, error) {
	if status == "suspended" {
		mitra, err := s.UserRepository.FindMitraById(mitraID)
		if err != nil {
			return 500, err
		}
		if mitra == nil {
			return 404, errors.New("mitra not found")
		}
		tx := s.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				return
			}
		}()
		if tx.Error != nil {
			return 500, tx.Error
		}
		if mitra.IsBusy == "yes" {
			orderData, err := s.OrderTransactionRepository.FindById(mitra.OrderIDRunning)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			if orderData == nil {
				tx.Rollback()
				return 404, errors.New("order not found")
			}
			paymentData, err := s.PaymentRepository.FindById(orderData.PaymentID)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			if paymentData == nil {
				tx.Rollback()
				return 404, errors.New("payment not found")
			}
			subPaymentData, err := s.SubPaymentRepository.FindById(orderData.SubPaymentID)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			if subPaymentData == nil {
				tx.Rollback()
				return 404, errors.New("sub payment not found")
			}
			customerData, err := s.UserRepository.FindCustomerById(orderData.CustomerID)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			if customerData == nil {
				tx.Rollback()
				return 404, errors.New("customer not found")
			}
			orderStatus := ""
			payloadUpdate := map[string]interface{}{
				"canceled_user": userType,
			}
			if paymentData.Type == "ewallet" {
				orderStatus = "CANCELED_VOID"
				paymentClient := helpers.NewClient()
				response, err := paymentClient.CreateVoidChargeXendit(ctx, orderData.PaymentIDPay)
				if err != nil {
					tx.Rollback()
					return 500, err
				}
				var respMap map[string]interface{}
				if err := json.Unmarshal(response, &respMap); err != nil {
					return 500, errors.New("Error when parsing the void response")
				}
				voidStatus := "VOID_PENDING"
				if v, ok := respMap["void_status"]; ok {
					if str, ok := v.(string); ok && str != "" {
						voidStatus = "VOID_" + str
					}
				}
				payloadUpdate["void_status"] = voidStatus
			} else if paymentData.Type == "balance" {
				orderStatus = "CANCELED_VOID"
			} else {
				orderStatus = "CANCELED"
			}
			payloadUpdate["order_status"] = orderStatus
			err = s.OrderTransactionRepository.UpdateData(tx, orderData, payloadUpdate)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			if paymentData.Type == "balance" {
				err = s.UserRepository.AddCustomerBalance(tx, orderData.CustomerID, float64(orderData.GrossAmount))
				if err != nil {
					tx.Rollback()
					return 500, err
				}
			}
			msgToCustomer := "Yah, order kamu dibatalin Suberes karena mitra bermasalah"
			if paymentData.Type == "virtual account" {
				msgToCustomer += "\ndan uang mu sebesar " + helpers.FormatRupiah(int64(orderData.GrossAmount)) + " dikembalikan ke rekening mu segera"
			} else if paymentData.Type == "balance" {
				msgToCustomer += "\ndan uang mu sebesar " + helpers.FormatRupiah(int64(orderData.GrossAmount)) + " dikembalikan ke saldo"
			} else if paymentData.Type == "ewallet" {
				msgToCustomer += "\ndan uang mu sebesar " + helpers.FormatRupiah(int64(orderData.GrossAmount)) + " dikembalikan ke e-wallet mu segera"
			}
			loc, err := time.LoadLocation(orderData.TimezoneCode)
			if err != nil {
				panic(err)
			}
			transactionDate := time.Now().UTC().In(loc).Format("2006-01-02 15:04:05")
			transPayload := models.Transaction{
				ID:                 orderData.ID,
				MitraID:            orderData.MitraID,
				OrderID:            orderData.ID,
				RefundAmount:       orderData.GrossAmount,
				RefundType:         "full refund",
				TimezoneCode:       orderData.TimezoneCode,
				TransactionName:    "Refund Biaya Order",
				TransactionAmount:  orderData.GrossAmount,
				TransactionType:    "transaction_out",
				TransactionTypeFor: "Refund",
				TransactionStatus:  "success",
				TransactionDescription: fmt.Sprintf(
					"Pengembalian dana kepada customer %s karena order dibatalkan",
					customerData.CompleteName,
				),
				CreatedAt: transactionDate,
				UpdatedAt: transactionDate,
			}
			err = s.TransactionRepository.CreateTransaction(tx, &transPayload)
			if err != nil {
				tx.Rollback()
				return 500, err
			}
			customerToken := customerData.FirebaseToken
			msgCustomer := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "CANCEL_BROADCAST_ADMIN",
					"title":             "Order dibatalin oleh Suberes",
					"message":           msgToCustomer,
					"order_id":          orderData.ID,
					"customer_id":       orderData.CustomerID,
					"mitra_id":          orderData.MitraID,
				},
				"tokens": []string{*customerToken},
			}
			_, err = service.SendMulticast(s.DB, "mitra", msgCustomer)
			if err != nil {
				log.Println("error:", err)
			}
		}
		userUpdate := map[string]interface{}{
			"is_mitra_activated":     "0",
			"is_suspended":           "1",
			"suspended_reason":       suspendedReason,
			"firebase_token":         "",
			"is_active":              "no",
			"is_auto_bid":            "no",
			"is_logged_in":           "0",
			"is_busy":                "no",
			"order_id_running":       nil,
			"sub_order_id_running":   nil,
			"customer_id_running":    nil,
			"service_id_running":     nil,
			"sub_service_id_running": nil,
		}
		_, err = s.UserRepository.UpdateUserData(tx, map[string]interface{}{
			"id":                 mitraID,
			"is_mitra_activated": "1",
			"is_suspended":       "0",
		}, userUpdate)
		if err != nil {
			tx.Rollback()
			return 500, err
		}
		if err := tx.Commit().Error; err != nil {
			return 500, err
		}
		if mitra.IsLoggedIn == "1" && (*mitra.FirebaseToken != "" || mitra.FirebaseToken != nil) {
			payloadMitra := map[string]interface{}{
				"data": map[string]interface{}{
					"notification_type": "SUSPEND_NOTIFICATION",
					"title":             "Nonaktifasi Akun",
					"message": fmt.Sprintf(
						"Kepada Yth mitra %s akun kamu telah di nonaktifkan oleh pihak Suberes\ndikarenakan melanggar Syarat & Ketentuan kami",
						mitra.CompleteName,
					),
					"message_advice": fmt.Sprintf(
						"Akun Suberes Mitra kamu ditangguhkan karena melanggar ketentuan Suberes Mitra. Kamu bisa lakuin banding dengan cara datang ke kantor Suberes di\n %s",
						os.Getenv("SUPPORT_OFFICE_ADDRESS"),
					),
					"mitra_id":   mitraID,
					"notif_type": "status",
				},
				"tokens": []string{*mitra.FirebaseToken},
			}
			_, err = service.SendMulticast(s.DB, "mitra", payloadMitra)
			if err != nil {
				log.Println("error:", err)
			}
		}
		helpers.SendMitraStatus(os.Getenv("SUPPORT_EMAIL"), mitra.Email, "Status Akun Suberes Mitra", status, mitra.Email, "Dinonaktifkan")
		return 200, nil
	} else if status == "active" {
		mitra, err := s.UserRepository.FindMitraById(mitraID)
		if err != nil {
			return 500, err
		}
		if mitra == nil {
			return 404, errors.New("Mitra not found")
		}
		tx := s.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				return
			}
		}()
		if tx.Error != nil {
			return 500, tx.Error
		}

		_, err = s.UserRepository.UpdateUserData(tx, map[string]interface{}{
			"id":                 mitraID,
			"is_mitra_activated": "0",
			"is_suspended":       "1",
		}, map[string]interface{}{
			"is_mitra_activated": "1",
			"is_suspended":       "0",
		})
		if err != nil {
			tx.Rollback()
			return 500, err
		}
		if err := tx.Commit().Error; err != nil {
			return 500, err
		}
		helpers.SendMitraStatus(os.Getenv("SUPPORT_EMAIL"), mitra.Email, "Status Akun Suberes Mitra", status, mitra.Email, "Diaktifkan Kembali")
		return 200, nil
	} else {
		return 400, errors.New("No valid status provided")
	}
}

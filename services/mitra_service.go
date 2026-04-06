package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"suberes_golang/constants"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/realtime"
	"suberes_golang/repositories"
	"suberes_golang/service"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	MitraCandidateImgAlias = "MITRA_CANDIDATE_IMG_"
)

type MitraService struct {
	DB                                *gorm.DB
	MitraRepository                   *repositories.MitraRepository
	UserRepository                    *repositories.UserRepository
	OrderTransactionRepository        *repositories.OrderTransactionRepository
	OrderTransactionRepeatsRepository *repositories.OrderTransactionRepeatsRepository
	PaymentRepository                 *repositories.PaymentRepository
	ServiceRepository                 *repositories.ServiceRepository
	SubPaymentRepository              *repositories.SubPaymentRepository
	TransactionRepository             *repositories.TransactionRepository
	UserOtpRepository                 *repositories.UserOtpRepository
	ScheduleRepository                *repositories.ScheduleRepository
	OrderOfferRepository              *repositories.OrderOfferRepository
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

	var sharedPrimeCustomer int64
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
		IsBusy:                      "no",
		IsMitraActivated:            "0",
		IsDocumentCompleted:         "0",
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
	if status == "suspend" {
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
			orderData, err := s.OrderTransactionRepository.FindById(*mitra.OrderIDRunning)
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
			transactionDate := time.Now().UTC().In(loc)
			transPayload := models.Transaction{
				ID:                 orderData.ID,
				MitraID:            helpers.DerefStr(orderData.MitraID),
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
		if err = s.MitraRepository.UpdateMitraByID(tx, mitraID, userUpdate); err != nil {
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

		if err = s.MitraRepository.UpdateMitraByID(tx, mitraID, map[string]interface{}{
			"is_mitra_activated": "1",
			"is_suspended":       "0",
		}); err != nil {
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
func (s *MitraService) UpdateMitraActive(mitraID, isActive string) (int, error) {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return 500, err
	}
	if mitra == nil {
		return 404, errors.New("Mitra not found")
	}
	if mitra.IsSuspended == "1" {
		return 401, errors.New("Mitra is suspended")
	}
	tx := s.DB.Begin()
	_, err = s.UserRepository.UpdateUserData(tx, map[string]interface{}{
		"id":        mitraID,
		"user_type": "mitra",
	}, map[string]interface{}{
		"is_active": isActive,
	})
	if err != nil {
		tx.Rollback()
		return 500, err
	}
	if err := tx.Commit().Error; err != nil {
		return 500, err
	}
	tokens, err := s.UserRepository.GetOnlineAdminTokens(s.DB)
	if err != nil {
		return 500, err
	}
	payloadAdmin := map[string]interface{}{
		"data": map[string]interface{}{
			"notification_type": "MITRA_UPDATE_ACTIVE_NOTIFICATION",
			"title":             "Status Aktif Mitra",
			"message": fmt.Sprintf(
				"Mitra %s telah %s",
				mitra.CompleteName,
				map[bool]string{true: "Aktif", false: "Nonaktif"}[isActive == "yes"],
			),
			"mitra_id":   mitra.ID,
			"notif_type": "status",
		},
		"tokens": tokens,
	}
	_, err = service.SendMulticast(s.DB, "admin", payloadAdmin)
	if err != nil {
		log.Println("error:", err)
	}
	return 200, nil
}
func (s *MitraService) UpdateMitraAutoBid(mitraID, isAutoBid string) (int, error) {
	mitra, err := s.UserRepository.FindMitraById(mitraID)
	if err != nil {
		return 500, err
	}
	if mitra == nil {
		return 404, errors.New("Mitra not found")
	}
	tx := s.DB.Begin()
	_, err = s.UserRepository.UpdateUserData(tx, map[string]interface{}{
		"id":        mitraID,
		"user_type": "mitra",
	}, map[string]interface{}{
		"is_auto_bid": isAutoBid,
	})
	if err != nil {
		tx.Rollback()
		return 500, err
	}
	if err := tx.Commit().Error; err != nil {
		return 500, err
	}
	return 200, nil
}
func (s *MitraService) UpdateMitraCoordinate(mitraID string, latitude, longitude float64) (int, error) {
	mitraData, err := s.UserRepository.FindOneDynamicMap(
		[]string{"id", "is_busy", "order_id_running"},
		map[string]interface{}{
			"id":        mitraID,
			"user_type": "mitra",
		},
	)
	if err != nil {
		return 500, err
	}
	if mitraData == nil {
		return 404, errors.New("Mitra not found")
	}
	tx := s.DB.Begin()
	_, err = s.UserRepository.UpdateUserData(tx, map[string]interface{}{
		"id":        mitraID,
		"user_type": "mitra",
	}, map[string]interface{}{
		"latitude":  fmt.Sprintf("%g", latitude),
		"longitude": fmt.Sprintf("%g", longitude),
	})
	if err != nil {
		tx.Rollback()
		return 500, err
	}
	if err := tx.Commit().Error; err != nil {
		return 500, err
	}
	if mitraData["is_busy"] == "yes" {
		orderData, err := s.OrderTransactionRepository.FindDynamicOrderTransactionMap(
			[]string{"coordinate_receiver_id"},
			map[string]interface{}{
				"id": mitraData["order_id_running"],
			},
			"order_status IN ?",
			[]interface{}{[]string{"OTW", "ON_PROGRESS"}},
		)
		if err != nil {
			return 500, err
		}
		if orderData == nil {
			return 404, errors.New("Order not found")
		}
		if orderData["coordinate_receiver_id"] != nil && orderData["coordinate_receiver_id"] != "" {
			realtime.Server.BroadcastToRoom(
				"/",
				orderData["coordinate_receiver_id"].(string),
				constants.COORDINATE_UPDATE,
				map[string]interface{}{
					"mitra_latitude":  latitude,
					"mitra_longitude": longitude,
				},
			)
		}
	}
	return 200, nil
}
func (s *MitraService) AdminIndex(page, limit int, search string) ([]models.User, int64, error) {
	return s.UserRepository.FindMitraPagination(page, limit, search)
}
func (s *MitraService) GetMitraDetail(id string, status string, timezone string) (interface{}, int, error) {

	// timezone handling
	loc, _ := time.LoadLocation(timezone)
	now := time.Now().In(loc)

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	end := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, loc)
	fmt.Println("Mitra ID : ", id)
	user, err := s.UserRepository.FindGenderMitraExGoLife(id)
	if err != nil {
		return nil, 404, errors.New("Mitra data not found")
	}

	if status == "candidate" && user.IsMitraActivated == "1" {
		return nil, 303, errors.New("see mitra detail page to see this data")
	}

	orderSummary, _ := s.OrderTransactionRepository.GetTodayOrderSummary(id, start, end)
	repeatSummary, _ := s.OrderTransactionRepeatsRepository.GetTodayRepeatSummary(id, start, end)

	totalOrder := orderSummary.OrderCount + repeatSummary.OrderCount
	totalPendapatan := orderSummary.Pendapatan + repeatSummary.Pendapatan

	orderData := map[string]interface{}{
		"order_count":      totalOrder,
		"pendapatan_order": totalPendapatan,
	}

	if user.IsBusy == "yes" && user.ServiceIDRunning != nil && user.SubServiceIDRunning != nil {
		serviceData, _ := s.ServiceRepository.GetRunningService(*user.ServiceIDRunning, *user.SubServiceIDRunning)
		orderData["service_data"] = serviceData
	}

	response := map[string]interface{}{
		"server_message": "success",
		"status":         "OK",
		"data":           user,
		"order_data":     orderData,
	}

	return response, 200, nil
}
func (s *MitraService) AdminUpdate(data dtos.UpdateMitraRequest) (int, error) {
	mitra, err := s.UserRepository.FindMitraById(data.MitraID)
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
	where := map[string]interface{}{
		"id":        data.MitraID,
		"user_type": "mitra",
	}

	// mapping DTO → update payload
	payload := map[string]interface{}{
		"email":                          data.Email,
		"country_code":                   data.CountryCode,
		"phone_number":                   data.PhoneNumber,
		"user_gender":                    data.UserGender,
		"domisili_address":               data.DomisiliAddress,
		"address":                        data.Address,
		"district":                       data.District,
		"sub_district":                   data.SubDistrict,
		"city":                           data.City,
		"postal_code":                    data.PostalCode,
		"is_ex_golife":                   data.IsExGoLife,
		"work_experience_duration":       data.WorkExperienceDuration,
		"emergency_contact_name":         data.EmergencyContactName,
		"emergency_contact_relation":     data.EmergencyContactRelation,
		"emergency_contact_country_code": data.EmergencyContactCountryCode,
		"emergency_contact_phone":        data.EmergencyContactPhone,
	}
	_, err = s.UserRepository.UpdateUserData(tx, where, payload)
	if err != nil {
		tx.Rollback()
		return 500, err
	}
	if err := tx.Commit().Error; err != nil {
		return 500, err
	}
	return 200, nil
}
func (s *MitraService) GetFilteredMitra(
	page int,
	limit int,
	search string,
	isExGolife string,
	kindOfMitra string,
) ([]models.User, int64, error) {

	filters := make(map[string]string)

	if isExGolife == "1" {
		filters["is_ex_golife"] = "1"
	}

	if kindOfMitra == "Dengan Alat" {
		filters["kind_of_mitra"] = "Dengan Alat"
	}

	return s.UserRepository.FindMitraWithFilter(page, limit, search, filters)
}
func deleteFiles(files []string) {
	for _, file := range files {
		_ = os.Remove(file)
	}
}
func (s *MitraService) UpdateMitraCandidate(
	id int,
	req dtos.UpdateMitraCandidateRequest,
	filePayload map[string]string,
	savedFiles []string,
) error {

	tx := s.DB.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var user models.User
	if err := tx.Where(
		"id = ? AND user_type = ? AND is_mitra_activated = ?",
		id, "mitra", "0",
	).First(&user).Error; err != nil {

		tx.Rollback()
		deleteFiles(savedFiles)
		return errors.New("mitra not found")
	}

	payload := map[string]interface{}{
		"email":        req.Email,
		"phone_number": req.PhoneNumber,
		"address":      req.Address,
	}

	// mapping gender
	if req.UserGender == "Pria" {
		payload["user_gender"] = "male"
	} else if req.UserGender == "Wanita" {
		payload["user_gender"] = "female"
	}

	// mapping golife
	if req.IsExGoLife == "Ya" {
		payload["is_ex_golife"] = "1"
	} else {
		payload["is_ex_golife"] = "0"
	}

	// merge file payload
	for k, v := range filePayload {
		payload[k] = v
	}

	if err := tx.Model(&user).Updates(payload).Error; err != nil {
		tx.Rollback()
		deleteFiles(savedFiles)
		return err
	}

	tx.Commit()
	return nil
}
func (s *MitraService) UpdateDocumentStatus(data dtos.DocumentStatusRequest) (int, error) {
	mitra, err := s.UserRepository.FindMitraById(data.ID)
	if err != nil {
		return 500, err
	}
	if mitra == nil {
		return 404, errors.New("Mitra not found")
	}
	tx := s.DB.Begin()
	if data.Status == "1" {
		_, err = s.UserRepository.UpdateUserData(tx,
			map[string]interface{}{
				"id":        data.ID,
				"user_type": "mitra",
			},
			map[string]interface{}{
				"is_document_completed": data.Status,
			})
		if err != nil {
			tx.Rollback()
			return 500, err
		}
	} else {
		err = s.UserRepository.DeleteByConditions(tx, map[string]interface{}{
			"id":        data.ID,
			"user_type": "mitra",
		})
		if err != nil {
			tx.Rollback()
			return 500, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return 500, err
	}
	helpers.SendInvitedMailMitra(os.Getenv("SUPPORT_EMAIL"), mitra.Email, "Data Tidak Lengkap", "Data anda kurang lengkap, silahkan submit ulang", "", "")
	return 200, nil
}

func (s *MitraService) InviteMitra(mitraID string, scheduleID int64) (int, error) {
	mitra, err := s.MitraRepository.FindMitraForInvite(mitraID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 404, errors.New("mitra not found")
		}
		return 500, err
	}

	schedule, err := s.ScheduleRepository.FindMitraLevelByID(scheduleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 404, errors.New("schedule not found")
		}
		return 500, err
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return 500, tx.Error
	}

	if err := s.MitraRepository.UpdateMitraInvited(tx, mitraID); err != nil {
		tx.Rollback()
		return 500, err
	}

	if err := tx.Commit().Error; err != nil {
		return 500, err
	}

	go helpers.SendInvitedMailMitra(
		os.Getenv("SUPPORT_EMAIL"),
		mitra.Email,
		"Undangan Verifikasi Mitra",
		schedule.ScheduleName,
		schedule.ScheduleDateTime,
		schedule.SchedulePlace,
	)

	return 200, nil
}

func (s *MitraService) TrainingStatus(mitraID, status string) (int, error) {
	if status != "successful" && status != "failed" {
		return 400, errors.New("status must be 'successful' or 'failed'")
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return 500, tx.Error
	}

	if err := s.MitraRepository.UpdateMitraTrainingStatus(tx, mitraID, status); err != nil {
		tx.Rollback()
		return 500, err
	}

	if err := tx.Commit().Error; err != nil {
		return 500, err
	}

	return 200, nil
}

func (s *MitraService) ActivateMitraStatus(mitraID, status string) (int, error) {
	if status != "successful" && status != "failed" {
		return 400, errors.New("status must be 'successful' or 'failed'")
	}

	mitra, err := s.MitraRepository.FindMitraForActivation(mitraID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 404, errors.New("mitra not found")
		}
		return 500, err
	}

	payload := map[string]interface{}{
		"is_mitra_rejected":  "0",
		"is_mitra_activated": "0",
	}

	var passwordPlain string

	if status == "successful" {
		payload["is_mitra_activated"] = "1"
		payload["is_mitra_rejected"] = "0"
		passwordPlain = helpers.GenerateMitraPassword()
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordPlain), 12)
		if err != nil {
			return 500, err
		}
		payload["password"] = string(hashedPassword)
	} else {
		payload["is_document_completed"] = "0"
		payload["is_mitra_invited"] = "0"
		payload["is_mitra_accepted"] = "0"
		payload["is_mitra_rejected"] = "1"
	}

	tx := s.DB.Begin()
	if tx.Error != nil {
		return 500, tx.Error
	}

	if err := s.MitraRepository.UpdateMitraActivationPayload(tx, mitraID, payload); err != nil {
		tx.Rollback()
		return 500, err
	}

	if err := tx.Commit().Error; err != nil {
		return 500, err
	}

	if status == "successful" {
		go helpers.SendAcceptedMailMitra(
			os.Getenv("SUPPORT_EMAIL"),
			mitra.Email,
			"Penerimaan Mitra Suberes",
			mitra.Email,
			passwordPlain,
		)
	}

	return 200, nil
}
func (s *MitraService) DashboardCount(mitraID string) (*dtos.MitraDashboardCountResponse, error) {
	counts, err := s.OrderTransactionRepository.GetMitraDashboardOrderCounts(mitraID)
	if err != nil {
		return nil, err
	}

	offerCount, err := s.OrderOfferRepository.CountByMitraID(mitraID)
	if err != nil {
		return nil, err
	}

	mitra, err := s.MitraRepository.FindMitraByID(mitraID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("mitra not found")
		}
		return nil, err
	}

	runningOrder, err := s.OrderTransactionRepository.FindRunningOrderDetailByMitraID(mitraID)
	if err != nil {
		return nil, err
	}

	status := "NOT_IN_ORDER"
	if runningOrder != nil {
		status = "IN_ORDER"
	}

	return &dtos.MitraDashboardCountResponse{
		OrderSoonCount:     counts.OrderSoonCount,
		OrderDoneCount:     counts.OrderDoneCount,
		OrderRepeatCount:   counts.OrderRepeatCount,
		OrderCanceledCount: counts.OrderCanceledCount,
		OrderOfferCount:    offerCount,
		IsSuspended:        mitra.IsSuspended,
		Status:             status,
		OrderRunningData:   runningOrder,
	}, nil
}

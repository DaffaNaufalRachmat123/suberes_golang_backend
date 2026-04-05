package services

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/repositories"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PinService struct {
	PinRepo     *repositories.PinRepository
	UserOTPRepo *repositories.UserOtpRepository
	DB          *gorm.DB
}

// GetPublicKeys returns the RSA public keys used to encrypt PINs on the client.
func (s *PinService) GetPublicKeys(userID string) (*repositories.PinPublicKeys, error) {
	return s.PinRepo.FindPublicKeys(userID)
}

// GetPinStatus returns whether pins are set along with descriptive status texts.
func (s *PinService) GetPinStatus(userID string) (*repositories.PinStatusResult, error) {
	return s.PinRepo.GetPinStatus(userID)
}

// CheckPin verifies that the supplied (RSA-encrypted) pin matches the stored encrypted pin.
func (s *PinService) CheckPin(userID, pinType, encryptedPin string) error {
	user, err := s.PinRepo.FindCustomerWithPins(userID)
	if err != nil {
		return err
	}

	var privateKey, storedPin string
	switch pinType {
	case "pay":
		privateKey = user.PrivateKeyPayPin
		storedPin = user.PayPin
	case "disbursement":
		privateKey = user.PrivateKeyDisbursementPin
		storedPin = user.DisbursementPin
	default:
		return errors.New("invalid pin type")
	}

	decrypted, err := helpers.DecryptRSA(privateKey, encryptedPin)
	if err != nil {
		return fmt.Errorf("rsa decrypt failed: %w", err)
	}

	encrypted, err := helpers.EncryptPinCbc(string(decrypted))
	if err != nil {
		return fmt.Errorf("aes encrypt failed: %w", err)
	}

	if encrypted != storedPin {
		return errors.New("old PIN is different")
	}
	return nil
}

// RequestChangePin validates the current PIN, then generates and sends an OTP for the change flow.
// Returns the new UserOTP record on success.
func (s *PinService) RequestChangePin(userID, pinType, encryptedPin string) (*models.UserOTP, error) {
	// 1. Validate old pin first
	if err := s.CheckPin(userID, pinType, encryptedPin); err != nil {
		return nil, err
	}

	// 2. Generate OTP
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	otpCode := fmt.Sprintf("%06d", rng.Intn(900000)+100000)

	hashedOtp, err := bcrypt.GenerateFromPassword([]byte(otpCode), 12)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 3. Delete any existing OTP for this user
	if err := s.UserOTPRepo.DeleteByUserId(tx, userID, map[string]interface{}{"otp_type": "change_pin"}); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Create new OTP record
	user, err := s.PinRepo.FindCustomerWithPins(userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	newOtp := &models.UserOTP{
		UsersID:     user.ID,
		OTPCode:     string(hashedOtp),
		OTPType:     "change_pin",
		SessionTime: time.Now(),
	}
	if err := s.UserOTPRepo.Create(newOtp, tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 5. Send OTP via email (async)
	fmt.Printf("OTP Code (change pin): %s\n", otpCode)
	go helpers.SendOtpCodeMail(os.Getenv("SUPPORT_EMAIL"), user.Email, "Kode OTP Suberes Ganti PIN", otpCode)

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return newOtp, nil
}

// ValidateOtp verifies the OTP code entered by the user and replaces the bcrypt hash with AES-encrypted value.
// The userOtpID is the ID of the users_otps record. encryptedOtp is RSA-encrypted by the client.
func (s *PinService) ValidateOtp(userID string, pinType string, userOtpID int, encryptedOtp string) error {
	user, err := s.PinRepo.FindCustomerWithPins(userID)
	if err != nil {
		return err
	}

	var privateKey string
	switch pinType {
	case "pay":
		privateKey = user.PrivateKeyPayPin
	case "disbursement":
		privateKey = user.PrivateKeyDisbursementPin
	default:
		return errors.New("invalid pin type")
	}

	decryptedOtp, err := helpers.DecryptRSA(privateKey, encryptedOtp)
	if err != nil {
		return fmt.Errorf("rsa decrypt failed: %w", err)
	}

	otpRecord, err := s.PinRepo.FindOTPForPinChange(userOtpID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("invalid request")
		}
		return err
	}

	// Compare decrypted input against bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(otpRecord.OTPCode), decryptedOtp); err != nil {
		return errors.New("OTP Code is wrong")
	}

	// Replace hash with AES-encrypted value for the configure/pin step
	aesEncoded, err := helpers.EncryptPinCbc(string(decryptedOtp))
	if err != nil {
		return fmt.Errorf("aes encrypt failed: %w", err)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.PinRepo.UpdateOTPCode(tx, userOtpID, aesEncoded); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// ConfigurePin sets a new PIN (change_pin=0) or changes an existing one via OTP (change_pin=1).
// Returns the updated PinStatusResult on success.
func (s *PinService) ConfigurePin(userID, pinType, changePin, encryptedPin string, userOtpID int, encryptedOtp string) (*repositories.PinStatusResult, error) {
	user, err := s.PinRepo.FindCustomerWithPins(userID)
	if err != nil {
		return nil, err
	}

	var privateKey, dbField string
	switch pinType {
	case "pay":
		privateKey = user.PrivateKeyPayPin
		dbField = "pay_pin"
	case "disbursement":
		privateKey = user.PrivateKeyDisbursementPin
		dbField = "disbursement_pin"
	default:
		return nil, errors.New("invalid pin type")
	}

	decryptedPin, err := helpers.DecryptRSA(privateKey, encryptedPin)
	if err != nil {
		return nil, fmt.Errorf("rsa decrypt failed: %w", err)
	}

	encryptedPinValue, err := helpers.EncryptPinCbc(string(decryptedPin))
	if err != nil {
		return nil, fmt.Errorf("aes encrypt failed: %w", err)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if changePin == "0" {
		// First-time setup: simply store the encrypted pin
		if err := s.PinRepo.UpdatePin(tx, userID, dbField, encryptedPinValue); err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		// Change existing pin: verify OTP first
		decryptedOtp, err := helpers.DecryptRSA(privateKey, encryptedOtp)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("rsa decrypt otp failed: %w", err)
		}

		aesOtp, err := helpers.EncryptPinCbc(string(decryptedOtp))
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("aes encrypt otp failed: %w", err)
		}

		otpRecord, err := s.PinRepo.FindOTPForPinChange(userOtpID)
		if err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("invalid request")
			}
			return nil, err
		}

		if otpRecord.OTPCode != aesOtp {
			tx.Rollback()
			return nil, errors.New("invalid request")
		}

		// OTP is valid – delete it and update pin
		if err := s.PinRepo.DestroyOTP(tx, userOtpID); err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := s.PinRepo.UpdatePin(tx, userID, dbField, encryptedPinValue); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return s.PinRepo.GetPinStatus(userID)
}

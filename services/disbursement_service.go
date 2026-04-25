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
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var failureMsgMap = map[string]string{
	"INVALID_DESTINATION":          "Rekening tak ditemukan",
	"SWITCHING_NETWORK_ERROR":      "Jaringan switching sedang ada masalah",
	"UNKNOWN_BANK_NETWORK_ERROR":   "Pihak Bank/Ewallet menolak transaksi tanpa alasan",
	"TEMPORARY_BANK_NETWORK_ERROR": "Bank/Ewallet sedang dalam perbaikan",
	"REJECTED_BY_BANK":             "Pihak Bank/Ewallet menolak transaksi",
	"TRANSFER_ERROR":               "Tarik tunai gagal karena kesalahan fatal",
	"TEMPORARY_TRANSFER_ERROR":     "Tarik tunai gagal karena kesalahan sementara",
	"INSUFFICIENT_BALANCE":         "Saldo tidak cukup",
}

func resolveFailureMsg(code string) string {
	if msg, ok := failureMsgMap[code]; ok {
		return msg
	}
	return "Gagal tarik tunai, kesalahan tak diketahui"
}

type TransactionDetailResponse struct {
	Transaction *models.Transaction `json:"transaction"`
	Bank        *models.BankList    `json:"bank"`
	FailureMsg  string              `json:"failure_msg"`
}

type DisbursementService struct {
	TransactionRepo *repositories.TransactionRepository
	UserRepo        *repositories.UserRepository
	DB              *gorm.DB
}

func NewDisbursementService(db *gorm.DB) *DisbursementService {
	return &DisbursementService{
		TransactionRepo: &repositories.TransactionRepository{DB: db},
		UserRepo:        &repositories.UserRepository{DB: db},
		DB:              db,
	}
}

func (s *DisbursementService) findBankByID(id int) (*models.BankList, error) {
	var bank models.BankList
	err := s.DB.Where("id = ?", id).First(&bank).Error
	if err != nil {
		return nil, err
	}
	return &bank, nil
}

// HandleTopupCallback processes the Xendit VA topup callback.
func (s *DisbursementService) HandleTopupCallback(data *dtos.TopupCallbackPayload) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	trx, err := s.TransactionRepo.FindTopupTransactionByExternalIDForCallback(data.ID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("transaction not found: %w", err)
	}

	if data.Amount != trx.TransactionAmount {
		tx.Rollback()
		return errors.New("transaction amount not same")
	}

	netAmount := trx.TransactionAmount - trx.TransactionFee

	if trx.UserType == "mitra" {
		user, err := s.UserRepo.FindMitraById(trx.MitraID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("mitra not found: %w", err)
		}

		if err := s.UserRepo.UpdateUserBalance(tx, user.ID, netAmount); err != nil {
			tx.Rollback()
			return err
		}

		lastAmount := user.AccountBalance + trx.TransactionAmount
		if err := s.TransactionRepo.UpdateTopupSuccess(tx, data.ID, lastAmount); err != nil {
			tx.Rollback()
			return err
		}

		log.Printf("[FCM] TOPUP_NOTIFICATION to mitra %s: amount=%d, transaction_id=%s", user.ID, netAmount, trx.ID)

	} else if trx.UserType == "customer" {
		user, err := s.UserRepo.FindCustomerById(trx.CustomerID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("customer not found: %w", err)
		}

		if err := s.UserRepo.UpdateUserBalance(tx, user.ID, netAmount); err != nil {
			tx.Rollback()
			return err
		}

		lastAmount := user.AccountBalance + trx.TransactionAmount
		if err := s.TransactionRepo.UpdateTopupSuccess(tx, data.ID, lastAmount); err != nil {
			tx.Rollback()
			return err
		}

		log.Printf("[FCM] TOPUP_NOTIFICATION to customer %s: amount=%d, transaction_id=%s", user.ID, netAmount, trx.ID)
	}

	return tx.Commit().Error
}

// HandleDisbursementCallback processes the Xendit disbursement callback.
func (s *DisbursementService) HandleDisbursementCallback(data *dtos.DisbursementCallbackPayload) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return tx.Error
	}

	trx, err := s.TransactionRepo.FindPendingDisbursementByExternalID(data.ID)
	if err != nil {
		tx.Rollback()
		return nil // transaction not found is treated as no-op (return 200)
	}

	if trx.UserType == "mitra" {
		user, err := s.UserRepo.FindMitraById(trx.MitraID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("mitra not found: %w", err)
		}

		if data.Status == "FAILED" {
			refundAmount := trx.TransactionAmount - trx.TransactionFee
			if err := s.TransactionRepo.UpdateDisbursementFailure(tx, trx.ID, data.FailureCode, refundAmount); err != nil {
				tx.Rollback()
				return err
			}
			if err := s.UserRepo.UpdateUserBalance(tx, user.ID, refundAmount); err != nil {
				tx.Rollback()
				return err
			}
		} else if data.Status == "COMPLETED" {
			if err := s.TransactionRepo.UpdateDisbursementStatus(tx, trx.ID, "success"); err != nil {
				tx.Rollback()
				return err
			}
		}

		log.Printf("[FCM] DISBURSEMENT_STATUS to mitra %s: status=%s, transaction_id=%s", user.ID, data.Status, trx.ID)

	} else if trx.UserType == "customer" {
		user, err := s.UserRepo.FindCustomerById(trx.CustomerID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("customer not found: %w", err)
		}

		if data.Status == "FAILED" {
			refundAmount := trx.TransactionAmount - trx.TransactionFee
			if err := s.TransactionRepo.UpdateDisbursementFailure(tx, trx.ID, data.FailureCode, refundAmount); err != nil {
				tx.Rollback()
				return err
			}
			if err := s.UserRepo.UpdateUserBalance(tx, user.ID, refundAmount); err != nil {
				tx.Rollback()
				return err
			}
		} else if data.Status == "COMPLETED" {
			if err := s.TransactionRepo.UpdateDisbursementStatus(tx, trx.ID, "success"); err != nil {
				tx.Rollback()
				return err
			}
		}

		log.Printf("[FCM] DISBURSEMENT_STATUS to customer %s: status=%s, transaction_id=%s", user.ID, data.Status, trx.ID)
	}

	return tx.Commit().Error
}

// CreateMitraTopup creates a topup transaction for a mitra.
func (s *DisbursementService) CreateMitraTopup(mitraID string, req *dtos.TopupRequest) (string, string, error) {
	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		return "", "", fmt.Errorf("bank not found")
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil {
		return "", "", fmt.Errorf("mitra not found")
	}

	if req.Amount < int64(bank.MinTopup) {
		return "", "", fmt.Errorf("min transactions for topup is %d", bank.MinTopup)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return "", "", tx.Error
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                mitra.ID,
		BankName:               req.BankName,
		BankID:                 &req.BankID,
		BankCode:               req.BankCode,
		IdempotencyKey:         idempotencyKey,
		UserType:               "mitra",
		TransactionName:        "Top Up Customer",
		TransactionAmount:      req.Amount,
		TransactionFee:         req.TopupFee,
		LastAmount:             mitra.AccountBalance,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "Top Up Customer",
		TransactionFor:         "topup",
		TransactionStatus:      "pending",
		TransactionDescription: fmt.Sprintf("Top Up Saldo Customer %s Sebesar Rp %d", mitra.CompleteName, req.Amount),
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if bank.MethodType == "ewallet" && bank.Code != "" && bank.CanTopup == "1" && bank.CanDisbursement == "1" {
		externalIDCreated := fmt.Sprintf("Topup-%s", helpers.GenerateInvoice("")[0:6])
		redirectURL := fmt.Sprintf("%s/api/disbursement/topup_payment_status/%s", os.Getenv("DIRECT_EWALLET_XENDIT"), externalIDCreated)

		payload := map[string]interface{}{
			"reference_id":    externalIDCreated,
			"currency":        "IDR",
			"amount":          req.Amount,
			"checkout_method": "ONE_TIME_PAYMENT",
			"channel_code":    bank.Code,
			"channel_properties": map[string]string{
				"success_redirect_url": redirectURL,
			},
			"metadata": map[string]string{
				"branch_area": "PLUIT",
				"branch_city": "JAKARTA",
			},
		}

		client := helpers.NewClient()
		respBytes, err := client.CreateEwalletChargeXendit(context.Background(), payload)
		if err != nil {
			tx.Rollback()
			return "", "", fmt.Errorf("xendit ewallet charge failed: %w", err)
		}

		var xenditResp dtos.XenditEwalletChargeAPIResponse
		if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
			tx.Rollback()
			return "", "", fmt.Errorf("failed to parse xendit response: %w", err)
		}

		trx.ExternalID = xenditResp.ID
		trx.TopupID = externalIDCreated
		trx.MobileEwalletURL = xenditResp.Actions.MobileWebCheckoutURL
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		return "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return "", "", err
	}

	return trx.ID, idempotencyKey, nil
}

// CreateCustomerTopup creates a topup transaction for a customer.
func (s *DisbursementService) CreateCustomerTopup(customerID string, req *dtos.TopupRequest) (string, string, error) {
	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		return "", "", fmt.Errorf("bank not found")
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		return "", "", fmt.Errorf("customer not found")
	}

	if req.Amount < int64(bank.MinTopup) {
		return "", "", fmt.Errorf("min transactions for topup is %d", bank.MinTopup)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return "", "", tx.Error
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)
	totalAmount := req.Amount + req.TopupFee

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		CustomerID:             customer.ID,
		BankName:               req.BankName,
		BankCode:               req.BankCode,
		BankID:                 &req.BankID,
		IdempotencyKey:         idempotencyKey,
		UserType:               "customer",
		TransactionName:        "Top Up Customer",
		TransactionAmount:      totalAmount,
		TransactionFee:         req.TopupFee,
		LastAmount:             customer.AccountBalance,
		TransactionType:        "transaction_in",
		TransactionTypeFor:     "Top Up Customer",
		TransactionFor:         "topup",
		TransactionStatus:      "pending",
		TransactionDescription: fmt.Sprintf("Top Up Saldo Customer %s Sebesar Rp %d", customer.CompleteName, req.Amount),
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if bank.MethodType == "ewallet" && bank.Code != "" && bank.CanTopup == "1" && bank.CanDisbursement == "1" {
		externalIDCreated := fmt.Sprintf("Topup-%s", helpers.GenerateInvoice("")[0:6])
		redirectURL := fmt.Sprintf("%s/api/disbursement/topup_payment_status/%s", os.Getenv("DIRECT_EWALLET_XENDIT"), externalIDCreated)

		payload := map[string]interface{}{
			"reference_id":    externalIDCreated,
			"currency":        "IDR",
			"amount":          totalAmount,
			"checkout_method": "ONE_TIME_PAYMENT",
			"channel_code":    bank.Code,
			"channel_properties": map[string]string{
				"success_redirect_url": redirectURL,
			},
			"metadata": map[string]string{
				"branch_area": "PLUIT",
				"branch_city": "JAKARTA",
			},
		}

		client := helpers.NewClient()
		respBytes, err := client.CreateEwalletChargeXendit(context.Background(), payload)
		if err != nil {
			tx.Rollback()
			return "", "", fmt.Errorf("xendit ewallet charge failed: %w", err)
		}

		var xenditResp dtos.XenditEwalletChargeAPIResponse
		if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
			tx.Rollback()
			return "", "", fmt.Errorf("failed to parse xendit response: %w", err)
		}

		trx.ExternalID = xenditResp.ID
		trx.TopupID = externalIDCreated
		trx.MobileEwalletURL = xenditResp.Actions.MobileWebCheckoutURL
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		return "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return "", "", err
	}

	return trx.ID, idempotencyKey, nil
}

// GetMitraTransactions returns paginated disbursement+topup transactions for a mitra.
func (s *DisbursementService) GetMitraTransactions(mitraID string, page, limit int) ([]models.Transaction, int64, error) {
	return s.TransactionRepo.FindMitraDisburseTransactionsPaginated(mitraID, page, limit)
}

// GetCustomerTransactions returns paginated disbursement+topup transactions for a customer.
func (s *DisbursementService) GetCustomerTransactions(customerID string, page, limit int) ([]models.Transaction, int64, error) {
	return s.TransactionRepo.FindCustomerDisburseTransactionsPaginated(customerID, page, limit)
}

// GetMitraTransactionDetail returns a single mitra transaction detail with bank data and failure message.
func (s *DisbursementService) GetMitraTransactionDetail(id, mitraID, idempotencyKey string) (*TransactionDetailResponse, error) {
	trx, err := s.TransactionRepo.FindMitraTransactionDetail(id, mitraID, idempotencyKey)
	if err != nil {
		return nil, err
	}

	var bank *models.BankList
	if trx.BankID != nil && *trx.BankID != 0 {
		bank, _ = s.findBankByID(*trx.BankID)
	}

	return &TransactionDetailResponse{
		Transaction: trx,
		Bank:        bank,
		FailureMsg:  resolveFailureMsg(trx.FailureCode),
	}, nil
}

// GetCustomerTransactionDetail returns a single customer transaction detail with bank data and failure message.
func (s *DisbursementService) GetCustomerTransactionDetail(id, customerID string) (*TransactionDetailResponse, error) {
	trx, err := s.TransactionRepo.FindCustomerTransactionDetail(id, customerID)
	if err != nil {
		return nil, err
	}

	var bank *models.BankList
	if trx.BankID != nil && *trx.BankID != 0 {
		bank, _ = s.findBankByID(*trx.BankID)
	}

	return &TransactionDetailResponse{
		Transaction: trx,
		Bank:        bank,
		FailureMsg:  resolveFailureMsg(trx.FailureCode),
	}, nil
}

// CreateMitraDisburse creates a disbursement transaction for a mitra.
func (s *DisbursementService) CreateMitraDisburse(mitraID string, req *dtos.DisburseRequest) (string, string, error) {
	if req.Amount < 6000 {
		return "", "", fmt.Errorf("amount less than Rp 6.000")
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil {
		return "", "", fmt.Errorf("mitra not found")
	}

	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		return "", "", fmt.Errorf("bank data not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(mitra.Password), []byte(req.Password)); err != nil {
		return "", "", fmt.Errorf("password not match")
	}

	totalAmount := req.Amount + int64(bank.DisbursementFee)

	externalID := fmt.Sprintf("Disbursement-%d", rand.Int63n(100000)+1)
	xenditPayload := map[string]interface{}{
		"external_id":         externalID,
		"amount":              totalAmount,
		"bank_code":           bank.DisbursementCode,
		"account_holder_name": req.AccountHolderName,
		"account_number":      req.AccountNumber,
		"description":         req.Description,
	}

	client := helpers.NewClient()
	respBytes, err := client.CreateDisbursementChargeXendit(context.Background(), xenditPayload)
	if err != nil {
		return "", "", fmt.Errorf("xendit disbursement failed: %w", err)
	}

	var xenditResp dtos.XenditDisbursementAPIResponse
	if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
		return "", "", fmt.Errorf("failed to parse xendit response: %w", err)
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return "", "", tx.Error
	}

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                mitraID,
		DisbursementID:         xenditResp.ExternalID,
		ExternalID:             xenditResp.ID,
		IdempotencyKey:         idempotencyKey,
		UserType:               "mitra",
		AccountOwnerName:       req.AccountHolderName,
		BankID:                 &req.BankID,
		BankName:               bank.Name,
		BankCode:               bank.DisbursementCode,
		AccountNumber:          req.AccountNumber,
		TransactionName:        "Disbursement Mitra",
		TransactionAmount:      totalAmount,
		TransactionFee:         int64(bank.DisbursementFee),
		LastAmount:             mitra.AccountBalance - totalAmount,
		TransactionType:        "transaction_out",
		TransactionTypeFor:     "Disbursement Mitra",
		TransactionFor:         "disbursement",
		TransactionStatus:      "pending",
		TransactionDescription: req.Description,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		return "", "", err
	}

	if err := s.UserRepo.DeductUserBalance(tx, mitraID, totalAmount); err != nil {
		tx.Rollback()
		return "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return "", "", err
	}

	return trx.ID, idempotencyKey, nil
}

// CreateCustomerDisburse creates a disbursement transaction for a customer.
func (s *DisbursementService) CreateCustomerDisburse(customerID string, req *dtos.DisburseCustomerRequest) (string, string, string, error) {
	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		return "", "", "", fmt.Errorf("bank data or ewallet data not found")
	}

	if req.Amount < int64(bank.MinDisbursement) {
		return "", "", "", fmt.Errorf("amount less than minimum disbursement")
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		return "", "", "", fmt.Errorf("customer not found")
	}

	if customer.DisbursementPin != "" {
		if req.Pin == "" {
			return "", "", "", fmt.Errorf("pin required")
		}
		decrypted, err := helpers.DecryptRSA(customer.PrivateKeyDisbursementPin, req.Pin)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to decrypt pin: %w", err)
		}
		encrypted, err := helpers.EncryptPinCbc(string(decrypted))
		if err != nil {
			return "", "", "", fmt.Errorf("failed to verify pin: %w", err)
		}
		if encrypted != customer.DisbursementPin {
			return "", "", "", fmt.Errorf("Unauthorized , PIN not match")
		}
	}

	totalAmount := req.Amount + int64(bank.DisbursementFee)
	externalID := fmt.Sprintf("Disbursement-%d", rand.Int63n(100000)+1)
	xenditPayload := map[string]interface{}{
		"external_id":         externalID,
		"amount":              totalAmount,
		"bank_code":           bank.DisbursementCode,
		"account_holder_name": req.AccountHolderName,
		"account_number":      req.AccountNumber,
		"description":         req.Description,
	}

	client := helpers.NewClient()
	respBytes, err := client.CreateDisbursementChargeXendit(context.Background(), xenditPayload)
	if err != nil {
		return "", "", "", fmt.Errorf("xendit disbursement failed: %w", err)
	}

	var xenditResp dtos.XenditDisbursementAPIResponse
	if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
		return "", "", "", fmt.Errorf("failed to parse xendit response: %w", err)
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return "", "", "", tx.Error
	}

	if err := s.UserRepo.DeductUserBalance(tx, customerID, totalAmount); err != nil {
		tx.Rollback()
		return "", "", "", err
	}

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		CustomerID:             customerID,
		DisbursementID:         xenditResp.ExternalID,
		ExternalID:             xenditResp.ID,
		IdempotencyKey:         idempotencyKey,
		UserType:               "customer",
		AccountOwnerName:       req.AccountHolderName,
		BankID:                 &req.BankID,
		BankName:               bank.Name,
		BankCode:               bank.DisbursementCode,
		AccountNumber:          req.AccountNumber,
		TransactionName:        "Disbursement Mitra",
		TransactionAmount:      totalAmount,
		TransactionFee:         int64(bank.DisbursementFee),
		LastAmount:             customer.AccountBalance - totalAmount,
		TransactionType:        "transaction_out",
		TransactionTypeFor:     "Disbursement Mitra",
		TransactionFor:         "disbursement",
		TransactionStatus:      "pending",
		TransactionDescription: req.Description,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		return "", "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		return "", "", "", err
	}

	return trx.ID, trx.ExternalID, idempotencyKey, nil
}

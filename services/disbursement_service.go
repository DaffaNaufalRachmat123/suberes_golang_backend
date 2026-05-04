package services

import (
	"context"
	"encoding/json"
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
		log.Printf("[DisbursementService] findBankByID error: %v", err)
		return nil, err
	}
	log.Printf("[DisbursementService] findBankByID response: %+v", bank)
	return &bank, nil
}

// CreateMitraTopup creates a topup transaction for a mitra.
func (s *DisbursementService) CreateMitraTopup(mitraID string, req *dtos.TopupRequest) (string, string, error) {
	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		log.Printf("[DisbursementService] CreateMitraTopup error: bank not found: %v", err)
		return "", "", fmt.Errorf("bank not found")
	}

	mitra, err := s.UserRepo.FindMitraById(mitraID)
	if err != nil {
		log.Printf("[DisbursementService] CreateMitraTopup error: mitra not found: %v", err)
		return "", "", fmt.Errorf("mitra not found")
	}

	if req.Amount < int64(bank.MinTopup) {
		log.Printf("[DisbursementService] CreateMitraTopup error: min transactions for topup is %d", bank.MinTopup)
		return "", "", fmt.Errorf("min transactions for topup is %d", bank.MinTopup)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		log.Printf("[DisbursementService] CreateMitraTopup error: tx begin: %v", tx.Error)
		return "", "", tx.Error
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		MitraID:                &mitra.ID,
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
		externalIDCreated := helpers.GenerateInvoice("Topup")
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
			log.Printf("[DisbursementService] CreateMitraTopup error: xendit ewallet charge failed: %v", err)
			return "", "", fmt.Errorf("xendit ewallet charge failed: %w", err)
		}

		var xenditResp dtos.XenditEwalletChargeAPIResponse
		if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateMitraTopup error: failed to parse xendit response: %v", err)
			return "", "", fmt.Errorf("failed to parse xendit response: %w", err)
		}
		log.Printf("[DisbursementService] CreateMitraTopup Xendit ewallet response: %s", string(respBytes))
		trx.ExternalID = xenditResp.ID
		trx.TopupID = externalIDCreated
		trx.MobileEwalletURL = xenditResp.Actions.MobileWebCheckoutURL
	} else if bank.MethodType == "bank" && bank.DisbursementCode != "" && bank.CanTopup == "1" {
		externalIDCreated := helpers.GenerateInvoice("Topup")

		vaPayload := map[string]interface{}{
			"external_id":     externalIDCreated,
			"bank_code":       bank.DisbursementCode,
			"name":            mitra.CompleteName,
			"expected_amount": req.Amount,
			"is_closed":       true,
			"is_single_use":   true,
		}

		client := helpers.NewClient()
		respBytes, err := client.CreateVirtualAccount(context.Background(), vaPayload)
		if err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateMitraTopup error: xendit VA creation failed: %v", err)
			return "", "", fmt.Errorf("xendit VA creation failed: %w", err)
		}

		var vaResp dtos.XenditVAAPIResponse
		if err := json.Unmarshal(respBytes, &vaResp); err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateMitraTopup error: failed to parse xendit VA response: %v", err)
			return "", "", fmt.Errorf("failed to parse xendit VA response: %w", err)
		}
		log.Printf("[DisbursementService] CreateMitraTopup Xendit VA response: %s", string(respBytes))
		trx.ExternalID = vaResp.ExternalID
		trx.TopupID = externalIDCreated
		trx.AccountNumber = vaResp.AccountNumber
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		log.Printf("[DisbursementService] CreateMitraTopup error: create transaction: %v", err)
		return "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[DisbursementService] CreateMitraTopup error: commit: %v", err)
		return "", "", err
	}

	log.Printf("[DisbursementService] CreateMitraTopup response: trxID=%s, idempotencyKey=%s", trx.ID, idempotencyKey)
	return trx.ID, idempotencyKey, nil
}

// CreateCustomerTopup creates a topup transaction for a customer.
func (s *DisbursementService) CreateCustomerTopup(customerID string, req *dtos.TopupRequest) (string, string, error) {
	bank, err := s.findBankByID(req.BankID)
	if err != nil {
		log.Printf("[DisbursementService] CreateCustomerTopup error: bank not found: %v", err)
		return "", "", fmt.Errorf("bank not found")
	}

	customer, err := s.UserRepo.FindCustomerById(customerID)
	if err != nil {
		log.Printf("[DisbursementService] CreateCustomerTopup error: customer not found: %v", err)
		return "", "", fmt.Errorf("customer not found")
	}

	if req.Amount < int64(bank.MinTopup) {
		log.Printf("[DisbursementService] CreateCustomerTopup error: min transactions for topup is %d", bank.MinTopup)
		return "", "", fmt.Errorf("min transactions for topup is %d", bank.MinTopup)
	}

	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		log.Printf("[DisbursementService] CreateCustomerTopup error: tx begin: %v", tx.Error)
		return "", "", tx.Error
	}

	idempotencyKey := fmt.Sprintf("%d", rand.Int63n(1000000)+1)
	totalAmount := req.Amount + req.TopupFee

	trx := &models.Transaction{
		ID:                     uuid.New().String(),
		CustomerID:             &customer.ID,
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
		externalIDCreated := helpers.GenerateInvoice("Topup")
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
			log.Printf("[DisbursementService] CreateCustomerTopup error: xendit ewallet charge failed: %v", err)
			return "", "", fmt.Errorf("xendit ewallet charge failed: %w", err)
		}

		var xenditResp dtos.XenditEwalletChargeAPIResponse
		if err := json.Unmarshal(respBytes, &xenditResp); err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateCustomerTopup error: failed to parse xendit response: %v", err)
			return "", "", fmt.Errorf("failed to parse xendit response: %w", err)
		}
		log.Printf("[DisbursementService] CreateCustomerTopup Xendit ewallet response: %s", string(respBytes))
		trx.ExternalID = xenditResp.ID
		trx.TopupID = externalIDCreated
		trx.MobileEwalletURL = xenditResp.Actions.MobileWebCheckoutURL
	} else if bank.MethodType == "bank" && bank.DisbursementCode != "" && bank.CanTopup == "1" {
		externalIDCreated := helpers.GenerateInvoice("Topup")

		vaPayload := map[string]interface{}{
			"external_id":     externalIDCreated,
			"bank_code":       bank.DisbursementCode,
			"name":            customer.CompleteName,
			"expected_amount": totalAmount,
			"is_single_use":   true,
		}

		client := helpers.NewClient()
		respBytes, err := client.CreateVirtualAccount(context.Background(), vaPayload)
		if err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateCustomerTopup error: xendit VA creation failed: %v", err)
			return "", "", fmt.Errorf("xendit VA creation failed: %w", err)
		}

		var vaResp dtos.XenditVAAPIResponse
		if err := json.Unmarshal(respBytes, &vaResp); err != nil {
			tx.Rollback()
			log.Printf("[DisbursementService] CreateCustomerTopup error: failed to parse xendit VA response: %v", err)
			return "", "", fmt.Errorf("failed to parse xendit VA response: %w", err)
		}
		log.Printf("[DisbursementService] CreateCustomerTopup Xendit VA response: %s", string(respBytes))
		trx.ExternalID = vaResp.ID
		trx.TopupID = externalIDCreated
		trx.AccountNumber = vaResp.AccountNumber
	}

	if err := s.TransactionRepo.CreateTransaction(tx, trx); err != nil {
		tx.Rollback()
		log.Printf("[DisbursementService] CreateCustomerTopup error: create transaction: %v", err)
		return "", "", err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("[DisbursementService] CreateCustomerTopup error: commit: %v", err)
		return "", "", err
	}

	log.Printf("[DisbursementService] CreateCustomerTopup response: trxID=%s, idempotencyKey=%s", trx.ID, idempotencyKey)
	return trx.ID, idempotencyKey, nil
}

// TopupPaymentStatusResponse is the response for GET /topup_payment_status/:topup_id.
type TopupPaymentStatusResponse struct {
	TopupID           string `json:"topup_id"`
	TransactionAmount int64  `json:"transaction_amount"`
	TransactionStatus string `json:"transaction_status"`
	Title             string `json:"title"`
	Description       string `json:"description"`
}

// GetTopupPaymentStatus returns the topup transaction status for the payment status page.
func (s *DisbursementService) GetTopupPaymentStatus(topupID string) (*TopupPaymentStatusResponse, error) {
	trx, err := s.TransactionRepo.FindByTopupID(topupID)
	if err != nil {
		log.Printf("[DisbursementService] GetTopupPaymentStatus error: %v", err)
		return nil, err
	}

	var statusText, statusDesc, fundDesc string
	switch trx.TransactionStatus {
	case "success":
		statusText = "Berhasil"
		statusDesc = "Berhasil"
		fundDesc = "Dana langsung masuk ke Saldo Suberes mu"
	case "pending":
		statusText = "sedang diproses"
		statusDesc = "Diproses"
		fundDesc = "Dana sedang diproses masuk ke Saldo Suberes mu"
	default:
		statusText = "Gagal"
		statusDesc = "Gagal"
		fundDesc = "Kami akan langsung balikin uang nya ke E-Wallet yang kamu pakai untuk bayar topup ini"
	}

	amountFormatted := fmt.Sprintf("Rp. %d", trx.TransactionAmount)
	title := fmt.Sprintf("Pembayaran TopUp\nsebesar %s\n%s", amountFormatted, statusText)
	description := fmt.Sprintf("Pembayaran TopUp dengan ID Transaksi %s sebesar %s %s. %s",
		trx.TopupID, amountFormatted, statusDesc, fundDesc)

	resp := &TopupPaymentStatusResponse{
		TopupID:           trx.TopupID,
		TransactionAmount: trx.TransactionAmount,
		TransactionStatus: trx.TransactionStatus,
		Title:             title,
		Description:       description,
	}
	log.Printf("[DisbursementService] GetTopupPaymentStatus response: %+v", resp)
	return resp, nil
}

// GetMitraTransactions returns paginated disbursement+topup transactions for a mitra.
func (s *DisbursementService) GetMitraTransactions(mitraID string, page, limit int) ([]models.Transaction, int64, error) {
	txs, total, err := s.TransactionRepo.FindMitraDisburseTransactionsPaginated(mitraID, page, limit)
	if err != nil {
		log.Printf("[DisbursementService] GetMitraTransactions error: %v", err)
	}
	log.Printf("[DisbursementService] GetMitraTransactions response: count=%d, total=%d", len(txs), total)
	return txs, total, err
}

// GetCustomerTransactions returns paginated disbursement+topup transactions for a customer.
func (s *DisbursementService) GetCustomerTransactions(customerID string, page, limit int) ([]models.Transaction, int64, error) {
	txs, total, err := s.TransactionRepo.FindCustomerDisburseTransactionsPaginated(customerID, page, limit)
	if err != nil {
		log.Printf("[DisbursementService] GetCustomerTransactions error: %v", err)
	}
	log.Printf("[DisbursementService] GetCustomerTransactions response: count=%d, total=%d", len(txs), total)
	return txs, total, err
}

// GetMitraTransactionDetail returns a single mitra transaction detail with bank data and failure message.
func (s *DisbursementService) GetMitraTransactionDetail(id, mitraID, idempotencyKey string) (*TransactionDetailResponse, error) {
	trx, err := s.TransactionRepo.FindMitraTransactionDetail(id, mitraID, idempotencyKey)
	if err != nil {
		log.Printf("[DisbursementService] GetMitraTransactionDetail error: %v", err)
		return nil, err
	}

	var bank *models.BankList
	if trx.BankID != nil && *trx.BankID != 0 {
		bank, _ = s.findBankByID(*trx.BankID)
	}

	resp := &TransactionDetailResponse{
		Transaction: trx,
		Bank:        bank,
		FailureMsg:  resolveFailureMsg(trx.FailureCode),
	}
	log.Printf("[DisbursementService] GetMitraTransactionDetail response: %+v", resp)
	return resp, nil
}

// GetCustomerTransactionDetail returns a single customer transaction detail with bank data and failure message.
func (s *DisbursementService) GetCustomerTransactionDetail(id, customerID string) (*TransactionDetailResponse, error) {
	trx, err := s.TransactionRepo.FindCustomerTransactionDetail(id, customerID)
	if err != nil {
		log.Printf("[DisbursementService] GetCustomerTransactionDetail error: %v", err)
		return nil, err
	}

	var bank *models.BankList
	if trx.BankID != nil && *trx.BankID != 0 {
		bank, _ = s.findBankByID(*trx.BankID)
	}

	resp := &TransactionDetailResponse{
		Transaction: trx,
		Bank:        bank,
		FailureMsg:  resolveFailureMsg(trx.FailureCode),
	}
	log.Printf("[DisbursementService] GetCustomerTransactionDetail response: %+v", resp)
	return resp, nil
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
	log.Printf("[DisbursementService] CreateMitraDisburse Xendit disbursement response: %s", string(respBytes))

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
		MitraID:                &mitraID,
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
	log.Printf("[DisbursementService] CreateCustomerDisburse Xendit disbursement response: %s", string(respBytes))

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
		CustomerID:             &customerID,
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

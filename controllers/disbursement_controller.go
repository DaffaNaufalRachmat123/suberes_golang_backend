package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DisbursementController struct {
	DisbursementService *services.DisbursementService
}

func NewDisbursementController(disbursementService *services.DisbursementService) *DisbursementController {
	return &DisbursementController{
		DisbursementService: disbursementService,
	}
}

// TopupCallback handles the Xendit VA topup webhook.
func (c *DisbursementController) TopupCallback(ctx *gin.Context) {
	var payload dtos.TopupCallbackPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid payload", "status": "failure"})
		return
	}

	if err := c.DisbursementService.HandleTopupCallback(&payload); err != nil {
		if err.Error() == "transaction amount not same" {
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.As(err, &gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "topup callback succeed", "status": "success"})
}

// DisbursementCallback handles the Xendit disbursement webhook.
func (c *DisbursementController) DisbursementCallback(ctx *gin.Context) {
	var payload dtos.DisbursementCallbackPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid payload", "status": "failure"})
		return
	}

	if err := c.DisbursementService.HandleDisbursementCallback(&payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "disbursement callback processed", "status": "success"})
}

// GetTopupPaymentStatus returns the topup transaction status (for the payment status page).
func (c *DisbursementController) GetTopupPaymentStatus(ctx *gin.Context) {
	topupID := ctx.Param("topup_id")

	result, err := c.DisbursementService.GetTopupPaymentStatus(topupID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "transaction not found", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// ValidateBank validates a bank account number against a bank code.
func (c *DisbursementController) ValidateBank(ctx *gin.Context) {
	var req dtos.ValidateBankRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure"})
		return
	}

	ewalletCodes := map[string]bool{"gopay": true, "shopeepay": true, "linkaja": true, "ovo": true}
	status := "valid"
	if ewalletCodes[req.BankCode] {
		status = "invalid"
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"account_number": req.AccountNumber,
			"bank_code":      req.BankCode,
			"account_holder": "DAFFA NAUFAL RACHMAT",
			"status":         status,
		},
	})
}

// CreateMitraTopup creates a topup transaction for a mitra.
func (c *DisbursementController) CreateMitraTopup(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	var req dtos.TopupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}

	transactionID, idempotencyKey, err := c.DisbursementService.CreateMitraTopup(mitraID, &req)
	if err != nil {
		switch err.Error() {
		case "bank not found":
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "disbursement with this method is not allowed", "status": "failure"})
		case "mitra not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "mitra not found", "status": "failure"})
		default:
			if len(err.Error()) > 20 && err.Error()[:20] == "min transactions for" {
				ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
			}
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message":  "top up created",
		"status":          "success",
		"transaction_id":  transactionID,
		"idempotency_key": idempotencyKey,
	})
}

// CreateCustomerTopup creates a topup transaction for a customer.
func (c *DisbursementController) CreateCustomerTopup(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	var req dtos.TopupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}

	transactionID, idempotencyKey, err := c.DisbursementService.CreateCustomerTopup(customerID, &req)
	if err != nil {
		switch err.Error() {
		case "bank not found":
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "disbursement with this method is not allowed", "status": "failure"})
		case "customer not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "customer not found", "status": "failure"})
		default:
			if len(err.Error()) > 20 && err.Error()[:20] == "min transactions for" {
				ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
			}
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message":  "top up created",
		"status":          "success",
		"transaction_id":  transactionID,
		"idempotency_key": idempotencyKey,
	})
}

// GetMitraTransactions returns paginated disbursement+topup transactions for a mitra.
func (c *DisbursementController) GetMitraTransactions(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	transactions, total, err := c.DisbursementService.GetMitraTransactions(mitraID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, transactions, len(transactions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// GetCustomerTransactions returns paginated disbursement+topup transactions for a customer.
func (c *DisbursementController) GetCustomerTransactions(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	transactions, total, err := c.DisbursementService.GetCustomerTransactions(customerID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, transactions, len(transactions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// GetMitraTransactionDetail returns a single mitra transaction detail.
func (c *DisbursementController) GetMitraTransactionDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	mitraID := ctx.Param("mitra_id")
	idempotencyKey := ctx.Param("idempotency_key")

	result, err := c.DisbursementService.GetMitraTransactionDetail(id, mitraID, idempotencyKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.APIErrorResponse(ctx, "transaction not found", http.StatusNotFound)
			return
		}
		helpers.APIErrorResponse(ctx, "Failed to get transaction", http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetCustomerTransactionDetail returns a single customer transaction detail.
func (c *DisbursementController) GetCustomerTransactionDetail(ctx *gin.Context) {
	id := ctx.Param("id")
	customerID := ctx.Param("customer_id")

	result, err := c.DisbursementService.GetCustomerTransactionDetail(id, customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			helpers.APIErrorResponse(ctx, "transaction not found", http.StatusNotFound)
			return
		}
		helpers.APIErrorResponse(ctx, "Failed to get transaction", http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// CreateMitraDisburse creates a disbursement transaction for a mitra.
func (c *DisbursementController) CreateMitraDisburse(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	var req dtos.DisburseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}

	transactionID, idempotencyKey, err := c.DisbursementService.CreateMitraDisburse(mitraID, &req)
	if err != nil {
		switch err.Error() {
		case "amount less than Rp 6.000":
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		case "mitra not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "mitra not found", "status": "failure"})
		case "bank data not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "bank data not found", "status": "failure"})
		case "password not match":
			ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": "password not match", "status": "failure"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message":  "disbursement created",
		"status":          "success",
		"id":              transactionID,
		"idempotency_key": idempotencyKey,
	})
}

// CreateCustomerDisburse creates a disbursement transaction for a customer.
func (c *DisbursementController) CreateCustomerDisburse(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	var req dtos.DisburseCustomerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure", "error": err.Error()})
		return
	}

	transactionID, externalID, idempotencyKey, err := c.DisbursementService.CreateCustomerDisburse(customerID, &req)
	if err != nil {
		switch err.Error() {
		case "bank data or ewallet data not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "bank data or ewallet data not found", "status": "failure"})
		case "amount less than minimum disbursement":
			ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "amount less", "status": "failure"})
		case "customer not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "customer not found", "status": "failure"})
		case "pin required", "Unauthorized , PIN not match":
			ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": err.Error(), "status": "failure"})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message":  "disbursement created",
		"status":          "success",
		"id":              transactionID,
		"external_id":     externalID,
		"idempotency_key": idempotencyKey,
	})
}

package controllers

import (
	"net/http"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

// WebhookController handles all inbound Xendit webhook HTTP requests.
// Each handler corresponds to exactly one Xendit event type / webhook URL.
type WebhookController struct {
	WebhookService *services.WebhookService
}

func NewWebhookController(webhookService *services.WebhookService) *WebhookController {
	return &WebhookController{WebhookService: webhookService}
}

// VACreate handles the Xendit "FVA Created/Updated" webhook.
// POST /api/webhook/va/create
//
// Routing (by external_id prefix):
//   - "Order-"        → update order PROCESSING_PAYMENT → WAITING_PAYMENT
//   - "Topup-"        → skip (200 OK)
//   - "Disbursement-" → skip (200 OK)
func (c *WebhookController) VACreate(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.WebhookService.HandleVACreate(body)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// VAPaid handles the Xendit "FVA Payment Received" webhook.
// POST /api/webhook/va/paid
//
// Routing (by external_id prefix):
//   - "Topup-"  → credit user balance
//   - otherwise → order WAITING_PAYMENT → FINDING_MITRA
func (c *WebhookController) VAPaid(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.WebhookService.HandleVAPaid(body)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Disbursement handles the Xendit disbursement result webhook.
// POST /api/webhook/disbursement
//
//   - status "COMPLETED" → mark transaction success
//   - status "FAILED"    → refund user balance, mark failed, send FCM
func (c *WebhookController) Disbursement(ctx *gin.Context) {
	var payload dtos.DisbursementCallbackPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid payload", "status": "failure"})
		return
	}
	if err := c.WebhookService.HandleDisbursement(&payload); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Ewallet handles the Xendit eWallet charge webhook (ewallet.capture / ewallet.void).
// POST /api/webhook/ewallet
//
//   - "ewallet.capture" → update order status, enqueue mitra search, send FCM
//   - "ewallet.void"    → mark order CANCELED_VOID, send FCM
func (c *WebhookController) Ewallet(ctx *gin.Context) {
	var payload dtos.XenditCallbackPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.WebhookService.HandleEwallet(payload)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

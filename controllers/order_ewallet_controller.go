package controllers

import (
	"net/http"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderEwalletController struct {
	OrderEwalletService *services.OrderEwalletService
}

func NewOrderEwalletController(orderEwalletService *services.OrderEwalletService) *OrderEwalletController {
	return &OrderEwalletController{OrderEwalletService: orderEwalletService}
}

// CallbackPaidPayment handles the Xendit ewallet.capture / ewallet.void webhook.
// POST /order_ewallet/callback
func (c *OrderEwalletController) CallbackPaidPayment(ctx *gin.Context) {
	var payload dtos.XenditCallbackPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.OrderEwalletService.CallbackPaidPayment(payload)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CallbackNotification is a no-op acknowledgement endpoint for Xendit ewallet notifications.
// POST /order_ewallet/notification/create
func (c *OrderEwalletController) CallbackNotification(ctx *gin.Context) {
	code, err := c.OrderEwalletService.CallbackNotification()
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CreateOrderEwallet creates a new ewallet order for a customer.
// POST /order_ewallet/create/:customer_id
func (c *OrderEwalletController) CreateOrderEwallet(ctx *gin.Context) {
	var req dtos.CreateOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	customerID := ctx.Param("customer_id")
	orderID, subID, custID, mitraID, code, err := c.OrderEwalletService.CreateOrderEwallet(customerID, req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"server_message": "order created",
		"status":         "success",
		"order_id":       orderID,
		"sub_id":         subID,
		"customer_id":    custID,
		"mitra_id":       mitraID,
	})
}

// AcceptOrderEwallet handles a mitra accepting a FINDING_MITRA ewallet order.
// POST /order_ewallet/accept
func (c *OrderEwalletController) AcceptOrderEwallet(ctx *gin.Context) {
	var req dtos.AcceptOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	code, response, err := c.OrderEwalletService.AcceptOrderEwallet(req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

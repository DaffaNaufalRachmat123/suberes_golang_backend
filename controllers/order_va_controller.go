package controllers

import (
	"net/http"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderVAController struct {
	OrderVAService *services.OrderVAService
}

func NewOrderVAController(orderVAService *services.OrderVAService) *OrderVAController {
	return &OrderVAController{OrderVAService: orderVAService}
}

// CallbackCreate handles the Xendit VA creation webhook (PROCESSING_PAYMENT → WAITING_PAYMENT).
// POST /order_va/notification/create
func (c *OrderVAController) CallbackCreate(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.OrderVAService.CallbackCreate(body)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CallbackPaidPayment handles the Xendit VA paid webhook (WAITING_PAYMENT → FINDING_MITRA).
// POST /order_va/notification/paid
func (c *OrderVAController) CallbackPaidPayment(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	code, err := c.OrderVAService.CallbackPaidPayment(body)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// CreateOrderVA creates a new virtual account order for a customer.
// POST /order_va/create/:customer_id
func (c *OrderVAController) CreateOrderVA(ctx *gin.Context) {
	var req dtos.CreateOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	customerID := ctx.Param("customer_id")
	orderID, subID, custID, mitraID, code, err := c.OrderVAService.CreateOrderVA(customerID, req)
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

// AcceptOrderVA handles a mitra accepting a FINDING_MITRA VA order.
// POST /order_va/accept
func (c *OrderVAController) AcceptOrderVA(ctx *gin.Context) {
	var req dtos.AcceptOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	code, response, err := c.OrderVAService.AcceptOrderVA(req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

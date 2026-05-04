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
	orderID, subID, custID, mitraID, code, err := c.OrderEwalletService.CreateOrderEwallet(ctx, customerID, req)
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

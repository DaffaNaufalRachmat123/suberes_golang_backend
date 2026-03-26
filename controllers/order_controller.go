package controllers

import (
	"net/http"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	OrderCashService *services.OrderCashService
}

func (c *OrderController) CreateOrderCash(ctx *gin.Context) {
	var req dtos.CreateOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	customerId := ctx.Param("customer_id")
	orderId, subId, customerID, mitraId, code, err := c.OrderCashService.CreateOrderCash(customerId, req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"server_message": "order created",
		"status":         "success",
		"order_id":       orderId,
		"sub_id":         subId,
		"customer_id":    customerID,
		"mitra_id":       mitraId,
	})
}

func (c *OrderController) AcceptOrderCash(ctx *gin.Context) {
	var req dtos.AcceptOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	code, response, err := c.OrderCashService.AcceptOrder(req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
	}
	ctx.JSON(http.StatusCreated, response)
}

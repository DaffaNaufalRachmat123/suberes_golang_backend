package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	OrderCashService *services.OrderCashService
	OrderService     *services.OrderService
}

func NewOrderController(orderCashService *services.OrderCashService, orderService *services.OrderService) *OrderController {
	return &OrderController{
		OrderCashService: orderCashService,
		OrderService:     orderService,
	}
}

func (c *OrderController) CreateOrderCash(ctx *gin.Context) {
	var req dtos.CreateOrderDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	customerId := ctx.Param("customer_id")
	orderId, subId, customerID, mitraId, code, err := c.OrderCashService.CreateOrderCash(customerId, req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
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
		return
	}
	code, response, err := c.OrderCashService.AcceptOrder(req)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

func (c *OrderController) FindAllByStatus(ctx *gin.Context) {
	status := ctx.Param("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.DefaultQuery("search", "")

	orders, total, err := c.OrderService.FindAllByStatusWithPagination(status, page, limit, search)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get orders", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, orders, len(orders), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

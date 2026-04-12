package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderOfferController struct {
	OrderOfferService *services.OrderOfferService
}

func NewOrderOfferController(orderOfferService *services.OrderOfferService) *OrderOfferController {
	return &OrderOfferController{OrderOfferService: orderOfferService}
}

// GetIncomingOrderList returns a paginated list of order offers for a given mitra.
// GET /order_offers/incoming_order_list/:mitra_id?page=1&limit=10
func (c *OrderOfferController) GetIncomingOrderList(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	offers, total, _, err := c.OrderOfferService.GetIncomingOrderList(mitraID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, offers, len(offers), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// GetIncomingOrder returns the detail of a single order offer for a mitra.
// GET /order_offers/incoming_order/:order_id/:mitra_id
func (c *OrderOfferController) GetIncomingOrder(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	mitraID := ctx.Param("mitra_id")

	offer, code, err := c.OrderOfferService.GetIncomingOrder(orderID, mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), code)
		return
	}

	helpers.APIResponse(ctx, "OK", http.StatusOK, offer)
}

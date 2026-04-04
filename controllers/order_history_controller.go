package controllers

import (
	"net/http"
	"strconv"

	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderHistoryController struct {
	OrderHistoryService *services.OrderHistoryService
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func parsePage(ctx *gin.Context) int {
	p, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if p < 1 {
		p = 1
	}
	return p
}

func parseLimit(ctx *gin.Context) int {
	l, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if l < 1 {
		l = 10
	}
	return l
}

// ── Order Canceleds ───────────────────────────────────────────────────────────

// GET /order_canceleds/index/date/:mitra_id
func (c *OrderHistoryController) GetCanceledDatesByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetCanceledDatesByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_canceleds/count/:mitra_id
func (c *OrderHistoryController) GetCanceledCountByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetCanceledCountByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_canceleds/index/:mitra_id/:order_time?page=&limit=&search=
func (c *OrderHistoryController) GetCanceledByMitraAndDate(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	orderTime := ctx.Param("order_time")
	search := ctx.DefaultQuery("search", "")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetCanceledByMitraAndDate(mitraID, orderTime, search, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_canceleds/list/customer/:customer_id?page=&limit=&start_date=&end_date=
func (c *OrderHistoryController) GetCanceledForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)
	startDate := ctx.DefaultQuery("start_date", "")
	endDate := ctx.DefaultQuery("end_date", "")

	data, total, err := c.OrderHistoryService.GetCanceledForCustomer(customerID, startDate, endDate, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// ── Order Dones ───────────────────────────────────────────────────────────────

// GET /order_dones/index/date/:mitra_id
func (c *OrderHistoryController) GetDoneDatesByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetDoneDatesByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_dones/count/:mitra_id
func (c *OrderHistoryController) GetDoneCountByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetDoneCountByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_dones/index/:mitra_id/:order_time?page=&limit=&search=
func (c *OrderHistoryController) GetDoneByMitraAndDate(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	orderTime := ctx.Param("order_time")
	search := ctx.DefaultQuery("search", "")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetDoneByMitraAndDate(mitraID, orderTime, search, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_dones/list/customer/:customer_id?page=&limit=&start_date=&end_date=
func (c *OrderHistoryController) GetDoneForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)
	startDate := ctx.DefaultQuery("start_date", "")
	endDate := ctx.DefaultQuery("end_date", "")

	data, total, err := c.OrderHistoryService.GetDoneForCustomer(customerID, startDate, endDate, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// ── Order Coming Soon ─────────────────────────────────────────────────────────

// GET /order_coming_soon/index/date/:mitra_id
func (c *OrderHistoryController) GetComingSoonDatesByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetComingSoonDatesByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_coming_soon/count/:mitra_id
func (c *OrderHistoryController) GetComingSoonCountByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetComingSoonCountByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_coming_soon/index/:mitra_id/:order_time?page=&limit=&search=
func (c *OrderHistoryController) GetComingSoonByMitraAndDate(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	orderTime := ctx.Param("order_time")
	search := ctx.DefaultQuery("search", "")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetComingSoonByMitraAndDate(mitraID, orderTime, search, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_coming_soon/list/customer/:customer_id?page=&limit=&start_date=&end_date=
func (c *OrderHistoryController) GetComingSoonForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)
	startDate := ctx.DefaultQuery("start_date", "")
	endDate := ctx.DefaultQuery("end_date", "")

	data, total, err := c.OrderHistoryService.GetComingSoonForCustomer(customerID, startDate, endDate, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// ── Order Repeat ──────────────────────────────────────────────────────────────

// GET /order_repeat/index/date/:mitra_id
func (c *OrderHistoryController) GetRepeatDatesByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetRepeatDatesByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_repeat/count/:mitra_id
func (c *OrderHistoryController) GetRepeatCountByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	data, err := c.OrderHistoryService.GetRepeatCountByMitra(mitraID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_repeat/index/:mitra_id/:order_time?page=&limit=&search=
func (c *OrderHistoryController) GetRepeatByMitraAndDate(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	orderTime := ctx.Param("order_time")
	search := ctx.DefaultQuery("search", "")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetRepeatByMitraAndDate(mitraID, orderTime, search, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_repeat/list/customer/:customer_id?page=&limit=
func (c *OrderHistoryController) GetRepeatForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetRepeatForCustomer(customerID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_repeat/list/repeat_list/:order_id/:mitra_id/:customer_id/:customer_timezone_code?page=&limit=
func (c *OrderHistoryController) GetRepeatList(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	mitraID := ctx.Param("mitra_id")
	customerID := ctx.Param("customer_id")
	// customer_timezone_code is accepted but countdown is computed server-side
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetRepeatList(orderID, mitraID, customerID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// ── Order Pending ─────────────────────────────────────────────────────────────

// GET /order_pending/count/:customer_id
func (c *OrderHistoryController) GetPendingCountByCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	data, err := c.OrderHistoryService.GetPendingCountByCustomer(customerID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /order_pending/index/mitra/:mitra_id?page=&limit=
func (c *OrderHistoryController) GetPendingByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetPendingByMitra(mitraID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// GET /order_pending/list/customer/:customer_id?page=&limit=
func (c *OrderHistoryController) GetPendingByCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetPendingByCustomer(customerID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

// ── Order Running ─────────────────────────────────────────────────────────────

// GET /order_running/list/customer/:customer_id?page=&limit=
func (c *OrderHistoryController) GetRunningForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	page := parsePage(ctx)
	limit := parseLimit(ctx)

	data, total, err := c.OrderHistoryService.GetRunningForCustomer(customerID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, data, len(data), page, limit, total))
}

package controllers

import (
	"net/http"
	"strconv"

	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type OrderTransactionController struct {
	OrderTransactionService *services.OrderTransactionService
}

// GET /orders/order_payment_status/:id_transaction
func (c *OrderTransactionController) GetPaymentStatus(ctx *gin.Context) {
	idTransaction := ctx.Param("id_transaction")
	order, err := c.OrderTransactionService.GetPaymentStatus(idTransaction)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusNotFound)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, order)
}

// GET /orders/index/admin/dashboard
func (c *OrderTransactionController) GetAdminDashboard(ctx *gin.Context) {
	stats, err := c.OrderTransactionService.GetAdminDashboard()
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, stats)
}

// GET /orders/timezone_code/:latitude/:longitude
func (c *OrderTransactionController) GetTimezoneCode(ctx *gin.Context) {
	lat := ctx.Param("latitude")
	lng := ctx.Param("longitude")
	tz, err := c.OrderTransactionService.GetTimezoneCode(lat, lng)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, tz)
}

// POST /orders/select_mitra/:order_id
func (c *OrderTransactionController) SelectMitra(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	var req services.SelectMitraRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	if err := c.OrderTransactionService.SelectMitra(orderID, req); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Mitras selected successfully", http.StatusOK, nil)
}

// GET /orders/index/selected_mitra/:order_id
func (c *OrderTransactionController) GetSelectedMitra(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	results, total, err := c.OrderTransactionService.GetSelectedMitra(orderID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	pagination := helpers.GetPaginationData(ctx, results, len(results), page, limit, total)
	helpers.APIResponse(ctx, "OK", http.StatusOK, pagination)
}

// GET /orders/order_detail_admin/:order_id
func (c *OrderTransactionController) GetAdminOrderDetail(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	order, err := c.OrderTransactionService.GetAdminOrderDetail(orderID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusNotFound)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, order)
}

// GET /orders/index/mitra/coming_soon/:mitra_id/:limit/:offset
func (c *OrderTransactionController) GetComingSoonOrdersForMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetComingSoonOrdersForMitra(mitraID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/running/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetRunningOrdersForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetRunningOrdersForCustomer(customerID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/canceled/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetCanceledOrdersForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetCanceledOrdersForCustomer(customerID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/done/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetDoneOrdersForCustomer(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetDoneOrdersForCustomer(customerID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/done/range_date/:start_date/:end_date/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetDoneOrdersRangeDate(ctx *gin.Context) {
	startDate := ctx.Param("start_date")
	endDate := ctx.Param("end_date")
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetDoneOrdersRangeDate(customerID, startDate, endDate, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/repeat/all/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetRepeatOrders(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetRepeatOrders(customerID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/index/repeat/all/search/:customer_id/:complete_name/:limit/:offset
func (c *OrderTransactionController) GetRepeatOrdersSearch(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	completeName := ctx.Param("complete_name")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetRepeatOrdersSearch(customerID, completeName, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/running_order/:order_id/:sub_id/:customer_id/:mitra_id/:type
func (c *OrderTransactionController) GetRunningOrderDetail(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")
	orderType := ctx.Param("type")

	subID, _ := strconv.Atoi(subIDStr)

	data, err := c.OrderTransactionService.GetRunningOrderDetail(orderID, subID, customerID, mitraID, orderType)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusNotFound)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /orders/index/virtual_account/all/:customer_id/:limit/:offset
func (c *OrderTransactionController) GetVirtualAccountOrders(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	limit, _ := strconv.Atoi(ctx.Param("limit"))
	offset, _ := strconv.Atoi(ctx.Param("offset"))
	if limit < 1 {
		limit = 10
	}

	orders, err := c.OrderTransactionService.GetVirtualAccountOrders(customerID, limit, offset)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, orders)
}

// GET /orders/order_detail/:order_id/:sub_id/:customer_id/:mitra_id/:load_all_repeat
func (c *OrderTransactionController) GetOrderDetailFull(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")
	loadAllRepeatStr := ctx.Param("load_all_repeat")

	subID, _ := strconv.Atoi(subIDStr)
	loadAllRepeat := loadAllRepeatStr == "true" || loadAllRepeatStr == "1"

	data, err := c.OrderTransactionService.GetOrderDetailFull(orderID, subID, customerID, mitraID, loadAllRepeat)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusNotFound)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// GET /orders/order_detail_customer/:order_id/:sub_id/:customer_id/:type/:is_load_repeat_list
func (c *OrderTransactionController) GetOrderDetailCustomer(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	orderType := ctx.Param("type")
	isLoadRepeatListStr := ctx.Param("is_load_repeat_list")
	mitraID := ctx.Param("mitra_id")

	subID, _ := strconv.Atoi(subIDStr)
	isLoadRepeatList := isLoadRepeatListStr == "true" || isLoadRepeatListStr == "1"

	data, err := c.OrderTransactionService.GetOrderDetailCustomer(orderID, subID, customerID, mitraID, orderType, isLoadRepeatList)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusNotFound)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, data)
}

// POST /orders/update_on_progress/:id/:customer_id/:mitra_id
func (c *OrderTransactionController) UpdateToOnProgress(ctx *gin.Context) {
	id := ctx.Param("id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	if err := c.OrderTransactionService.UpdateToOnProgress(id, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order updated to ON_PROGRESS", http.StatusOK, nil)
}

// POST /orders/update_on_progress/repeat/:id/:sub_id/:customer_id/:mitra_id
func (c *OrderTransactionController) UpdateToOnProgressRepeat(ctx *gin.Context) {
	id := ctx.Param("id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	subID, _ := strconv.Atoi(subIDStr)

	if err := c.OrderTransactionService.UpdateToOnProgressRepeat(id, subID, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Repeat order updated to ON_PROGRESS", http.StatusOK, nil)
}

// POST /orders/update_on_finish/:id/:customer_id/:mitra_id
func (c *OrderTransactionController) UpdateToFinish(ctx *gin.Context) {
	id := ctx.Param("id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	if err := c.OrderTransactionService.UpdateToFinish(id, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order updated to FINISH", http.StatusOK, nil)
}

// POST /orders/update_on_finish/repeat/:id/:sub_id/:customer_id/:mitra_id
func (c *OrderTransactionController) UpdateToFinishRepeat(ctx *gin.Context) {
	id := ctx.Param("id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	subID, _ := strconv.Atoi(subIDStr)

	if err := c.OrderTransactionService.UpdateToFinishRepeat(id, subID, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Repeat order updated to FINISH", http.StatusOK, nil)
}

// POST /orders/cancel_blast/:order_id
func (c *OrderTransactionController) CancelBlast(ctx *gin.Context) {
	orderID := ctx.Param("order_id")

	if err := c.OrderTransactionService.CancelBlast(orderID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order blast canceled", http.StatusOK, nil)
}

// POST /orders/rejected/:customer_id/:mitra_id/:service_id/:sub_service_id
func (c *OrderTransactionController) RejectOrder(ctx *gin.Context) {
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")
	serviceIDStr := ctx.Param("service_id")
	subServiceIDStr := ctx.Param("sub_service_id")

	serviceID, _ := strconv.Atoi(serviceIDStr)
	subServiceID, _ := strconv.Atoi(subServiceIDStr)

	if err := c.OrderTransactionService.RejectOrder(customerID, mitraID, serviceID, subServiceID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "Order rejected", http.StatusOK, nil)
}

// POST /orders/cancel/admin/:id
func (c *OrderTransactionController) AdminCancelOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	var req services.AdminCancelRequest
	_ = ctx.ShouldBindJSON(&req)

	if err := c.OrderTransactionService.AdminCancelOrder(id, req.CanceledReason); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order canceled by admin", http.StatusOK, nil)
}

// POST /orders/cancel/:id/:customer_id/:mitra_id/:canceled_user
func (c *OrderTransactionController) CancelOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")
	canceledUser := ctx.Param("canceled_user")

	var req services.CancelOrderRequest
	_ = ctx.ShouldBindJSON(&req)

	if err := c.OrderTransactionService.CancelOrder(id, customerID, mitraID, canceledUser, req.CanceledReason); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order canceled", http.StatusOK, nil)
}

// POST /orders/cancel/repeat/:id/:sub_id/:customer_id/:mitra_id/:canceled_user
func (c *OrderTransactionController) CancelRepeatOrder(ctx *gin.Context) {
	id := ctx.Param("id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")
	canceledUser := ctx.Param("canceled_user")

	subID, _ := strconv.Atoi(subIDStr)

	var req services.CancelOrderRequest
	_ = ctx.ShouldBindJSON(&req)

	if err := c.OrderTransactionService.CancelRepeatOrder(id, subID, customerID, mitraID, canceledUser, req.CanceledReason); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Repeat order canceled", http.StatusOK, nil)
}

// POST /orders/update/repeat/run_order/:order_id/:sub_id/:customer_id/:mitra_id
func (c *OrderTransactionController) StartRepeatRunOrder(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	subIDStr := ctx.Param("sub_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	subID, _ := strconv.Atoi(subIDStr)

	if err := c.OrderTransactionService.StartRepeatRunOrder(orderID, subID, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Repeat order started", http.StatusOK, nil)
}

// PUT /orders/update/run_order/:order_id/:customer_id/:mitra_id
func (c *OrderTransactionController) StartRunOrder(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	if err := c.OrderTransactionService.StartRunOrder(orderID, customerID, mitraID); err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusBadRequest)
		return
	}
	helpers.APIResponse(ctx, "Order started", http.StatusOK, nil)
}

// GET /orders/is_auto_bid/:order_id/:mitra_id
func (c *OrderTransactionController) IsAutoBid(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	mitraID := ctx.Param("mitra_id")

	offer, err := c.OrderTransactionService.IsAutoBid(orderID, mitraID)
	if err != nil {
		helpers.APIResponse(ctx, "Not found", http.StatusOK, gin.H{"is_auto_bid": false})
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, gin.H{"is_auto_bid": offer != nil, "offer": offer})
}

// GET /orders/directions/:order_id
func (c *OrderTransactionController) GetDirections(ctx *gin.Context) {
	orderID := ctx.Param("order_id")

	result, err := c.OrderTransactionService.GetDirections(orderID)
	if err != nil {
		helpers.APIErrorResponse(ctx, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.APIResponse(ctx, "OK", http.StatusOK, result)
}

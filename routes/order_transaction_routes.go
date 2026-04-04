package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderTransactionRoutes(r *gin.RouterGroup, controller *controllers.OrderTransactionController, db *gorm.DB) {
	order := r.Group("/orders")

	// ── Public routes (no auth required) ────────────────────────────────────────
	order.GET("/order_payment_status/:id_transaction", controller.GetPaymentStatus)
	order.GET("/timezone_code/:latitude/:longitude", controller.GetTimezoneCode)
	order.GET("/index/mitra/coming_soon/:mitra_id/:limit/:offset", controller.GetComingSoonOrdersForMitra)
	order.GET("/running_order/:order_id/:sub_id/:customer_id/:mitra_id/:type", controller.GetRunningOrderDetail)
	order.GET("/index/virtual_account/all/:customer_id/:limit/:offset", controller.GetVirtualAccountOrders)
	order.PUT("/update/run_order/:order_id/:customer_id/:mitra_id", controller.StartRunOrder)

	// ── Protected routes ─────────────────────────────────────────────────────────
	protected := order.Group("")
	protected.Use(middleware.AuthMiddleware(db))

	protectedRoutes := []helpers.ProtectedRoute{
		// Admin-only routes
		{
			Method:  "GET",
			Path:    "/index/admin/dashboard",
			Handler: controller.GetAdminDashboard,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/selected_mitra/:order_id",
			Handler: controller.GetSelectedMitra,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/order_detail_admin/:order_id",
			Handler: controller.GetAdminOrderDetail,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/cancel/admin/:id",
			Handler: controller.AdminCancelOrder,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/select_mitra/:order_id",
			Handler: controller.SelectMitra,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		// Customer routes
		{
			Method:  "GET",
			Path:    "/index/running/:customer_id/:limit/:offset",
			Handler: controller.GetRunningOrdersForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/canceled/:customer_id/:limit/:offset",
			Handler: controller.GetCanceledOrdersForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/done/:customer_id/:limit/:offset",
			Handler: controller.GetDoneOrdersForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/done/range_date/:start_date/:end_date/:customer_id/:limit/:offset",
			Handler: controller.GetDoneOrdersRangeDate,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/repeat/all/:customer_id/:limit/:offset",
			Handler: controller.GetRepeatOrders,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/repeat/all/search/:customer_id/:complete_name/:limit/:offset",
			Handler: controller.GetRepeatOrdersSearch,
			Roles:   []string{helpers.CustomerRole},
		},
		// Mitra routes
		{
			Method:  "GET",
			Path:    "/is_auto_bid/:order_id/:mitra_id",
			Handler: controller.IsAutoBid,
			Roles:   []string{helpers.MitraRole},
		},
		// Customer + Mitra shared routes
		{
			Method:  "GET",
			Path:    "/order_detail/:order_id/:sub_id/:customer_id/:mitra_id/:load_all_repeat",
			Handler: controller.GetOrderDetailFull,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/order_detail_customer/:order_id/:sub_id/:customer_id/:type/:is_load_repeat_list",
			Handler: controller.GetOrderDetailCustomer,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/update_on_progress/:id/:customer_id/:mitra_id",
			Handler: controller.UpdateToOnProgress,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/update_on_progress/repeat/:id/:sub_id/:customer_id/:mitra_id",
			Handler: controller.UpdateToOnProgressRepeat,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/update_on_finish/:id/:customer_id/:mitra_id",
			Handler: controller.UpdateToFinish,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/update_on_finish/repeat/:id/:sub_id/:customer_id/:mitra_id",
			Handler: controller.UpdateToFinishRepeat,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/cancel_blast/:order_id",
			Handler: controller.CancelBlast,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/rejected/:customer_id/:mitra_id/:service_id/:sub_service_id",
			Handler: controller.RejectOrder,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/cancel/:id/:customer_id/:mitra_id/:canceled_user",
			Handler: controller.CancelOrder,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/cancel/repeat/:id/:sub_id/:customer_id/:mitra_id/:canceled_user",
			Handler: controller.CancelRepeatOrder,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/update/repeat/run_order/:order_id/:sub_id/:customer_id/:mitra_id",
			Handler: controller.StartRepeatRunOrder,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		// All roles: directions
		{
			Method:  "GET",
			Path:    "/directions/:order_id",
			Handler: controller.GetDirections,
			Roles:   helpers.AllRole,
		},
	}

	helpers.RegisterProtectedRoutes(protected, protectedRoutes)
}

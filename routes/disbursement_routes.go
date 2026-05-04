package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DisbursementRoutes(r *gin.RouterGroup, controller *controllers.DisbursementController, db *gorm.DB) {
	disbursement := r.Group("/disbursement")

	// Public endpoint: topup payment status page (no auth)
	disbursement.GET("/topup_payment_status/:topup_id", controller.GetTopupPaymentStatus)

	// Auth-protected endpoints
	disbursement.Use(middleware.AuthMiddleware(db))

	protectedRoutes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/topup/:mitra_id",
			Handler: controller.CreateMitraTopup,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/topup/customer/:customer_id",
			Handler: controller.CreateCustomerTopup,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/validate",
			Handler: controller.ValidateBank,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/mitra/:mitra_id",
			Handler: controller.GetMitraTransactions,
			Roles:   []string{helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/customer/:customer_id",
			Handler: controller.GetCustomerTransactions,
			Roles:   []string{helpers.CustomerRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/transaction/mitra/:id/:mitra_id/:idempotency_key",
			Handler: controller.GetMitraTransactionDetail,
			Roles:   []string{helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/transaction/customer/:id/:customer_id",
			Handler: controller.GetCustomerTransactionDetail,
			Roles:   []string{helpers.CustomerRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create/:mitra_id",
			Handler: controller.CreateMitraDisburse,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "POST",
			Path:    "/create/customer/:customer_id",
			Handler: controller.CreateCustomerDisburse,
			Roles:   []string{helpers.CustomerRole},
		},
	}
	helpers.RegisterProtectedRoutes(disbursement, protectedRoutes)
}

package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TransactionRoutes(r *gin.RouterGroup, controller *controllers.TransactionController, db *gorm.DB) {
	transaction := r.Group("/transactions")
	transaction.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/admin",
			Handler: controller.FindAll,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/type/:mitra_id/:pendapatan_date",
			Handler: controller.GetTransactionTypes,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:transaction_for/:transaction_time",
			Handler: controller.FindAllByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/disbursement/:mitra_id",
			Handler: controller.FindDisbursementsByMitra,
			Roles:   []string{helpers.MitraRole},
		},
	}
	helpers.RegisterProtectedRoutes(transaction, routes)
}

package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BankListRoutes(r *gin.RouterGroup, controller *controllers.BankListController, db *gorm.DB) {
	bankList := r.Group("/bank")
	bankList.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/bank_list/topup",
			Handler: controller.GetTopupBanks,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "GET",
			Path:    "/bank_list",
			Handler: controller.GetDisbursementBanks,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "POST",
			Path:    "/bank_list/:admin_id",
			Handler: controller.BulkCreateBanks,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/ewallet_list/:admin_id",
			Handler: controller.BulkCreateEwallets,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/bank_ewallet_list/:id",
			Handler: controller.UpdateBankEwallet,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(bankList, routes)
}

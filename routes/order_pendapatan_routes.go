package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderPendapatanRoutes(r *gin.RouterGroup, controller *controllers.OrderTransactionController, db *gorm.DB) {
	group := r.Group("/order_pendapatan")
	protected := group.Group("")
	protected.Use(middleware.AuthMiddleware(db))

	protectedRoutes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:payment_id/:order_time",
			Handler: controller.GetOrderPendapatan,
			Roles:   []string{helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole}, // Allowed roles
		},
	}

	helpers.RegisterProtectedRoutes(protected, protectedRoutes)
}
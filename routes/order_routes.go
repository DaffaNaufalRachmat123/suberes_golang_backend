package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderRoutes(r *gin.RouterGroup, controller *controllers.OrderController, db *gorm.DB) {
	order := r.Group("/cash_order")
	order.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create/:customer_id",
			Handler: controller.CreateOrderCash,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/accept",
			Handler: controller.AcceptOrderCash,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/admin/:status",
			Handler: controller.FindAllByStatus,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(order, routes)
}

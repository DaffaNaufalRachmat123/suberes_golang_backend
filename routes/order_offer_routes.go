package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderOfferRoutes(r *gin.RouterGroup, controller *controllers.OrderOfferController, db *gorm.DB) {
	offers := r.Group("/order_offers")
	offers.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/incoming_order_list/:mitra_id",
			Handler: controller.GetIncomingOrderList,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/incoming_order/:order_id/:mitra_id",
			Handler: controller.GetIncomingOrder,
			Roles:   []string{helpers.MitraRole},
		},
	}
	helpers.RegisterProtectedRoutes(offers, routes)
}

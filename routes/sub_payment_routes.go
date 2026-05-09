package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SubPaymentRoutes(r *gin.RouterGroup, controller *controllers.SubPaymentController, db *gorm.DB) {
	subPayment := r.Group("/sub_payments")

	protected := subPayment.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.Index,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/detail/:id",
			Handler: controller.Detail,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/:id",
			Handler: controller.Update,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
	}

	helpers.RegisterProtectedRoutes(protected, routes)
}

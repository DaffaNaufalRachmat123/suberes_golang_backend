package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SubServiceRoutes(r *gin.RouterGroup, controller *controllers.SubServiceController, db *gorm.DB) {
	subservices := r.Group("/sub_service")
	protected := subservices.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/:id",
			Handler: controller.Update,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/:id",
			Handler: controller.Delete,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
	}

	helpers.RegisterProtectedRoutes(protected, routes)
}

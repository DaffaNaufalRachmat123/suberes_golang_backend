package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func LayananServiceRoutes(r *gin.RouterGroup, controller *controllers.LayananServiceController, db *gorm.DB) {
	layananService := r.Group("/layanan_service")
	layananService.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.Index,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/popular",
			Handler: controller.GetPopular,
			Roles:   []string{helpers.SuperAdminRole, helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/detail/:id",
			Handler: controller.GetByID,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/image/:id",
			Handler: controller.Update,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(layananService, routes)
}

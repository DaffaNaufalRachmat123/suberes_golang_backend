package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ServiceRoutes(r *gin.RouterGroup, controller *controllers.ServiceController, db *gorm.DB) {
	service := r.Group("/services")
	protected := service.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	service.GET("/detail/:id", controller.GetDetail)
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/:parent_id",
			Handler: controller.Index,
			Roles:   []string{helpers.CustomerRole, helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/search/:layanan_id",
			Handler: controller.SearchService,
			Roles:   []string{helpers.CustomerRole, helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/popular",
			Handler: controller.Popular,
			Roles:   []string{helpers.CustomerRole, helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/layanan_service/:id",
			Handler: controller.LayananServices,
			Roles:   []string{helpers.CustomerRole, helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/service_type/:id",
			Handler: controller.ServiceType,
			Roles:   []string{helpers.CustomerRole, helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/image",
			Handler: controller.UpdateImage,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update",
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

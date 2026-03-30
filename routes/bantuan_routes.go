package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BantuanRoutes(r *gin.RouterGroup, controller *controllers.BantuanController, db *gorm.DB) {
	bantuan := r.Group("/bantuan")
	bantuan.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/customer",
			Handler: controller.IndexCustomer,
			Roles:   []string{helpers.SuperAdminRole, helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/mitra",
			Handler: controller.IndexMitra,
			Roles:   []string{helpers.SuperAdminRole, helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/admin",
			Handler: controller.IndexAdmin,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/get/:id",
			Handler: controller.GetByID,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole, helpers.CustomerRole, helpers.MitraRole},
		},
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
	helpers.RegisterProtectedRoutes(bantuan, routes)
}

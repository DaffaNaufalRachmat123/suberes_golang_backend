package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PanduanRoutes(router *gin.RouterGroup, controller *controllers.PanduanController, db *gorm.DB) {
	panduan := router.Group("/guide")

	panduan.Use(middleware.AuthMiddleware(db))

	panduan.GET("/index/customer", controller.IndexCustomer)
	panduan.GET("/index/mitra", controller.IndexMitra)

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/admin",
			Handler: controller.IndexAdmin,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/detail/:id",
			Handler: controller.Detail,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/watching_count/:id",
			Handler: controller.UpdateWatchingCount,
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

	helpers.RegisterProtectedRoutes(panduan, routes)
}

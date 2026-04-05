package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ComplainRoutes(r *gin.RouterGroup, controller *controllers.ComplainController, db *gorm.DB) {
	complain := r.Group("/complains")
	protected := complain.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.IndexAdmin,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/customer",
			Handler: controller.IndexCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/index/mitra",
			Handler: controller.IndexMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/detail/:id",
			Handler: controller.Detail,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/:id/:status",
			Handler: controller.UpdateStatus,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/:id",
			Handler: controller.Remove,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

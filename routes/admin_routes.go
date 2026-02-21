package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminRoutes(r *gin.RouterGroup, controller *controllers.AdminController, db *gorm.DB) {
	admin := r.Group("/admin")
	admin.POST("/login", controller.Login)

	protected := admin.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/dashboard",
			Handler: controller.GetDashboard,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index_admin",
			Handler: controller.IndexAdmin,
			Roles:   []string{helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.CreateAdmin,
			Roles:   []string{helpers.SuperAdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update_admin_status/:admin_id",
			Handler: controller.UpdateAdminStatus,
			Roles:   []string{helpers.SuperAdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/admin/:admin_id",
			Handler: controller.RemoveAdmin,
			Roles:   []string{helpers.SuperAdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/logout",
			Handler: controller.Logout,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update_firebase_token/:id",
			Handler: controller.UpdateFirebaseToken,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ScheduleRoutes(router *gin.RouterGroup, controller *controllers.ScheduleController, db *gorm.DB) {
	schedules := router.Group("/schedules")
	schedules.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.Index,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
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
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/:id",
			Handler: controller.Update,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/:id/:password",
			Handler: controller.Delete,
			Roles:   []string{helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(schedules, routes)
}

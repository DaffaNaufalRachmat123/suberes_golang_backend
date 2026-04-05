package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewsRoutes(router *gin.RouterGroup, controller *controllers.NewsController, db *gorm.DB) {
	news := router.Group("/news")
	news.Use(middleware.AuthMiddleware(db))
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
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/detail/:id",
			Handler: controller.GetByID,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole, helpers.CustomerRole},
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
	helpers.RegisterProtectedRoutes(news, routes)
}

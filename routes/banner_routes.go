package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func BannerRoutes(r *gin.RouterGroup, controller *controllers.BannerController, db *gorm.DB) {
	banner := r.Group("/banners")
	banner.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.Index,
			Roles:   []string{helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/index/popular",
			Handler: controller.GetPopular,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/get/:id",
			Handler: controller.GetByID,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/image/:id",
			Handler: controller.Update,
			Roles:   []string{helpers.AdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/:id",
			Handler: controller.Delete,
			Roles:   []string{helpers.AdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(banner, routes)
}

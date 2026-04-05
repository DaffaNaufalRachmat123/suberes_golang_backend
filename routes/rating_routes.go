package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RatingRoutes(r *gin.RouterGroup, controller *controllers.RatingController, db *gorm.DB) {
	rating := r.Group("/ratings")
	protected := rating.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/mitra/:id/:limit/:offset",
			Handler: controller.GetMitraRatings,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
		{
			Method:  "GET",
			Path:    "/customer/:id/:limit/:offset",
			Handler: controller.GetCustomerRatings,
			Roles:   []string{helpers.CustomerRole, helpers.MitraRole, helpers.AdminRole, helpers.SuperAdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

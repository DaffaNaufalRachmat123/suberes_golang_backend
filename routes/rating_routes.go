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
			Roles:   helpers.AllRole,
		},
		{
			Method:  "GET",
			Path:    "/customer/:id/:limit/:offset",
			Handler: controller.GetCustomerRatings,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "GET",
			Path:    "/mitra/home/:mitra_id",
			Handler: controller.GetMitraRatingHome,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "GET",
			Path:    "/mitra/list/:mitra_id",
			Handler: controller.GetMitraRatingsPaginated,
			Roles:   helpers.AllRole,
		},
		{
			Method:  "POST",
			Path:    "/create_to_mitra/:order_id/:customer_id/:mitra_id/:rating",
			Handler: controller.CreateRatingToMitra,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/create_to_customer/:order_id/:customer_id/:mitra_id/:rating",
			Handler: controller.CreateRatingToCustomer,
			Roles:   []string{helpers.MitraRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

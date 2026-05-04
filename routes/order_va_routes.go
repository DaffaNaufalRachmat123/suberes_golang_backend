package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderVARoutes(r *gin.RouterGroup, controller *controllers.OrderVAController, db *gorm.DB) {
	va := r.Group("/va_order")

	// ── Protected routes ───────────────────────────────────────────────────────────
	protected := va.Group("")
	protected.Use(middleware.AuthMiddleware(db))

	protectedRoutes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create/:customer_id",
			Handler: controller.CreateOrderVA,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/accept",
			Handler: controller.AcceptOrderVA,
			Roles:   []string{helpers.MitraRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, protectedRoutes)
}

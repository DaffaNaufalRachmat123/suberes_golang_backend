package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderEwalletRoutes(r *gin.RouterGroup, controller *controllers.OrderEwalletController, db *gorm.DB) {
	ewallet := r.Group("/ewallet_order")

	// ── Xendit webhook endpoints (callback token required) ─────────────────────────
	xenditWebhook := ewallet.Group("")
	xenditWebhook.Use(middleware.XenditCallbackTokenMiddleware())
	{
		xenditWebhook.POST("/callback", controller.CallbackPaidPayment)
		xenditWebhook.POST("/notification/create", controller.CallbackNotification)
	}

	// ── Protected routes ──────────────────────────────────────────────────────────
	protected := ewallet.Group("")
	protected.Use(middleware.AuthMiddleware(db))

	protectedRoutes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create/:customer_id",
			Handler: controller.CreateOrderEwallet,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/accept",
			Handler: controller.AcceptOrderEwallet,
			Roles:   []string{helpers.MitraRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, protectedRoutes)
}

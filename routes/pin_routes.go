package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PinRoutes(r *gin.RouterGroup, controller *controllers.PinController, db *gorm.DB) {
	pin := r.Group("/pins")
	protected := pin.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/customer/public_key",
			Handler: controller.GetPublicKey,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/customer/pin_status",
			Handler: controller.GetPinStatus,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/customer/pin_check",
			Handler: controller.PinCheck,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/customer/request_change_pin",
			Handler: controller.RequestChangePin,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/customer/otp_validate",
			Handler: controller.OtpValidate,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "POST",
			Path:    "/customer/configure/pin",
			Handler: controller.ConfigurePin,
			Roles:   []string{helpers.CustomerRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

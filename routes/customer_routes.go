package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CustomerRoutes(r *gin.RouterGroup, controller *controllers.CustomerController, db *gorm.DB) {
	customer := r.Group("/customer")
	customer.POST("/login/email", controller.LoginByEmail)
	customer.POST("/otp_validator/mail", controller.OtpValidatorMail)
	customer.POST("/register", controller.Register)
	protected := customer.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "PUT",
			Path:    "/update_firebase_token",
			Handler: controller.UpdateFirebaseToken,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/profile",
			Handler: controller.GetCustomerProfile,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "PUT",
			Path:    "/change_phone_mail",
			Handler: controller.ChangePhoneMail,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "PUT",
			Path:    "/otp_update_phone_mail",
			Handler: controller.OtpUpdatePhoneMail,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "PUT",
			Path:    "/update_user_profile",
			Handler: controller.UpdateUserProfile,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "DELETE",
			Path:    "/logout",
			Handler: controller.UserLogout,
			Roles:   []string{helpers.CustomerRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

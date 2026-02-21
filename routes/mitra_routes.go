package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MitraRoutes(router *gin.RouterGroup, controller *controllers.MitraController, db *gorm.DB) {
	mitraRoutes := router.Group("/api/mitra")
	mitraRoutes.POST("/login", controller.Login)
	mitraRoutes.POST("/register", controller.Register)
	mitraRoutes.PUT("/change_forgot_password", controller.ChangeForgotPassword)
	mitraRoutes.PUT("/request_forgot_password/:email", controller.RequestForgotPassword)
	mitraRoutes.PUT("/otp_validator_forgot_password", controller.OTPValidatorForgotPassword)

	protected := mitraRoutes.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/update_firebase_token",
			Handler: controller.Profile,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/get_email_password/:id",
			Handler: controller.GetEmailPassword,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "PUT",
			Path:    "/change_password/:id",
			Handler: controller.ChangePassword,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "PUT",
			Path:    "/change_email/:id",
			Handler: controller.ChangeEmail,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "DELETE",
			Path:    "/logout/:id",
			Handler: controller.Logout,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "PUT",
			Path:    "/update_firebase_token/:id/:firebase_token",
			Handler: controller.UpdateFirebaseToken,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "PUT",
			Path:    "/update_mitra_status/:id/:status",
			Handler: controller.UpdateMitraStatus,
			Roles:   []string{helpers.AdminRole},
		},
	}
	helpers.RegisterProtectedRoutes(protected, routes)
}

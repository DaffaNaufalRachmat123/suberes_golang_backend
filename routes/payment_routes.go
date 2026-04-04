package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PaymentRoutes(r *gin.RouterGroup, controller *controllers.PaymentController, db *gorm.DB) {
	payment := r.Group("/payments")

	// GET /payments/index is public – mirrors the original JS route that has no
	// passport.authenticate middleware, so customers and guests can fetch payment methods.
	payment.GET("/index", controller.Index)

	// All mutating routes require JWT + Admin / SuperAdmin role.
	protected := payment.Group("/")
	protected.Use(middleware.AuthMiddleware(db))
	routes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/image/:id",
			Handler: controller.UpdateWithImage,
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
	helpers.RegisterProtectedRoutes(protected, routes)
}

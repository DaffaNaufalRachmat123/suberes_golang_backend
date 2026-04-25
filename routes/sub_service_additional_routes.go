package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SubServiceAdditionalRoutes(r *gin.RouterGroup, controller *controllers.SubServiceAdditionalController, db *gorm.DB) {
	subServiceAdditional := r.Group("/sub_service_additional")
	protected := subServiceAdditional.Group("/")
	protected.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "POST",
			Path:    "/create",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
	}

	helpers.RegisterProtectedRoutes(protected, routes)
}

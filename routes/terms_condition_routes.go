package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TermsConditionRoutes(router *gin.RouterGroup, controller *controllers.TermsConditionController, db *gorm.DB) {
	termsCondition := router.Group("/toc")

	termsCondition.GET("/user/:toc_type/:toc_user_type", controller.GetByTypeAndUserType)

	termsCondition.Use(middleware.AuthMiddleware(db))

	routes := []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index",
			Handler: controller.Index,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "GET",
			Path:    "/detail/:id",
			Handler: controller.Detail,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "POST",
			Path:    "/create/:force",
			Handler: controller.Create,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/:id/:force",
			Handler: controller.Update,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "PUT",
			Path:    "/update/status/:id",
			Handler: controller.UpdateStatus,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
		{
			Method:  "DELETE",
			Path:    "/remove/:id",
			Handler: controller.Delete,
			Roles:   []string{helpers.SuperAdminRole, helpers.AdminRole},
		},
	}

	helpers.RegisterProtectedRoutes(termsCondition, routes)
}

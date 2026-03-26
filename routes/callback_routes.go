package routes

import (
	"suberes_golang/controllers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupCallbackRoutes(router *gin.RouterGroup) {
	xenditController := controllers.NewXenditController()

	xenditGroup := router.Group("/xendit")
	xenditGroup.Use(middleware.XenditCallbackTokenMiddleware())
	{
		xenditGroup.POST("/ewallet", xenditController.EwalletCallback)
	}
	router.POST("/notification", xenditController.CallbackNotification)
}

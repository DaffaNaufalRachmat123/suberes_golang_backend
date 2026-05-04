package routes

import (
	"suberes_golang/controllers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
)

// WebhookRoutes registers all Xendit inbound webhook endpoints under /api/webhook/.
//
// All routes require the Xendit callback token (x-callback-token header).
//
// Route summary:
//
//	POST /api/webhook/va/create    → FVA created/updated (order activation, topup/disbursement skipped)
//	POST /api/webhook/va/paid      → FVA payment received (order paid, topup credited)
//	POST /api/webhook/disbursement → Disbursement completed/failed (refund or success)
//	POST /api/webhook/ewallet      → eWallet charge event (capture or void)
func WebhookRoutes(r *gin.RouterGroup, controller *controllers.WebhookController) {
	wh := r.Group("/webhook")
	wh.Use(middleware.XenditCallbackTokenMiddleware())
	{
		// Virtual Account events
		va := wh.Group("/va")
		{
			va.POST("/create", controller.VACreate)
			va.POST("/paid", controller.VAPaid)
		}

		// Disbursement (outgoing bank transfer) result
		wh.POST("/disbursement", controller.Disbursement)

		// eWallet charge events
		wh.POST("/ewallet", controller.Ewallet)
	}
}

package routes

import (
	"suberes_golang/controllers"
	"suberes_golang/helpers"
	middleware "suberes_golang/middlewares"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OrderHistoryRoutes(r *gin.RouterGroup, c *controllers.OrderHistoryController, db *gorm.DB) {

	// ── Order Canceleds ───────────────────────────────────────────────────────
	canceled := r.Group("/order_canceleds")

	// Public (no auth in original JS)
	canceled.GET("/index/date/:mitra_id", c.GetCanceledDatesByMitra)
	canceled.GET("/count/:mitra_id", c.GetCanceledCountByMitra)
	canceled.GET("/index/:mitra_id/:order_time", c.GetCanceledByMitraAndDate)

	// Protected — customer only
	canceledProtected := canceled.Group("")
	canceledProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(canceledProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetCanceledForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Dones ───────────────────────────────────────────────────────────
	done := r.Group("/order_dones")

	// Public
	done.GET("/index/date/:mitra_id", c.GetDoneDatesByMitra)
	done.GET("/count/:mitra_id", c.GetDoneCountByMitra)
	done.GET("/index/:mitra_id/:order_time", c.GetDoneByMitraAndDate)

	// Protected — customer only
	doneProtected := done.Group("")
	doneProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(doneProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetDoneForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Coming Soon ─────────────────────────────────────────────────────
	comingSoon := r.Group("/order_coming_soon")

	// Public
	comingSoon.GET("/index/date/:mitra_id", c.GetComingSoonDatesByMitra)
	comingSoon.GET("/count/:mitra_id", c.GetComingSoonCountByMitra)
	comingSoon.GET("/index/:mitra_id/:order_time", c.GetComingSoonByMitraAndDate)

	// Protected — customer only
	comingSoonProtected := comingSoon.Group("")
	comingSoonProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(comingSoonProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetComingSoonForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Repeat ──────────────────────────────────────────────────────────
	repeat := r.Group("/order_repeat")

	// Public
	repeat.GET("/index/date/:mitra_id", c.GetRepeatDatesByMitra)
	repeat.GET("/count/:mitra_id", c.GetRepeatCountByMitra)
	repeat.GET("/index/:mitra_id/:order_time", c.GetRepeatByMitraAndDate)

	// Protected — customer only
	repeatProtected := repeat.Group("")
	repeatProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(repeatProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetRepeatForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/list/repeat_list/:order_id/:mitra_id/:customer_id/:customer_timezone_code",
			Handler: c.GetRepeatList,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Pending ─────────────────────────────────────────────────────────
	pending := r.Group("/order_pending")

	// Public (no auth in original JS)
	pending.GET("/count/:customer_id", c.GetPendingCountByCustomer)
	pending.GET("/list/customer/:customer_id", c.GetPendingByCustomer)

	// Protected — mitra only
	pendingProtected := pending.Group("")
	pendingProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(pendingProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/mitra/:mitra_id",
			Handler: c.GetPendingByMitra,
			Roles:   []string{helpers.MitraRole},
		},
	})

	// ── Order Running ─────────────────────────────────────────────────────────
	running := r.Group("/order_running")

	// Protected — customer only
	runningProtected := running.Group("")
	runningProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(runningProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetRunningForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})
}

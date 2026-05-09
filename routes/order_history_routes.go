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
	canceled := r.Group("/order_canceled")

	// Protected — customer only
	canceledProtected := canceled.Group("")
	canceledProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(canceledProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/date/:mitra_id",
			Handler: c.GetCanceledDatesByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/count/:mitra_id",
			Handler: c.GetCanceledCountByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:order_time",
			Handler: c.GetCanceledByMitraAndDate,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetCanceledForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Dones ───────────────────────────────────────────────────────────
	done := r.Group("/order_dones")

	// Protected — customer only
	doneProtected := done.Group("")
	doneProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(doneProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/date/:mitra_id",
			Handler: c.GetDoneDatesByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/count/:mitra_id",
			Handler: c.GetDoneCountByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:order_time",
			Handler: c.GetDoneByMitraAndDate,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetDoneForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Coming Soon ─────────────────────────────────────────────────────
	comingSoon := r.Group("/order_coming_soon")

	// Protected — customer only
	comingSoonProtected := comingSoon.Group("")
	comingSoonProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(comingSoonProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/date/:mitra_id",
			Handler: c.GetComingSoonDatesByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/count/:mitra_id",
			Handler: c.GetComingSoonCountByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:order_time",
			Handler: c.GetComingSoonByMitraAndDate,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetComingSoonForCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
	})

	// ── Order Repeat ──────────────────────────────────────────────────────────
	repeat := r.Group("/order_repeat")

	// Protected — customer only
	repeatProtected := repeat.Group("")
	repeatProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(repeatProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/index/date/:mitra_id",
			Handler: c.GetRepeatDatesByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/count/:mitra_id",
			Handler: c.GetRepeatCountByMitra,
			Roles:   []string{helpers.MitraRole},
		},
		{
			Method:  "GET",
			Path:    "/index/:mitra_id/:order_time",
			Handler: c.GetRepeatByMitraAndDate,
			Roles:   []string{helpers.MitraRole},
		},
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

	// Protected — mitra only
	pendingProtected := pending.Group("")
	pendingProtected.Use(middleware.AuthMiddleware(db))
	helpers.RegisterProtectedRoutes(pendingProtected, []helpers.ProtectedRoute{
		{
			Method:  "GET",
			Path:    "/count/:customer_id",
			Handler: c.GetPendingCountByCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
		{
			Method:  "GET",
			Path:    "/list/customer/:customer_id",
			Handler: c.GetPendingByCustomer,
			Roles:   []string{helpers.CustomerRole},
		},
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

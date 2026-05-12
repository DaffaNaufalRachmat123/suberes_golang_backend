package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"suberes_golang/config"
	"suberes_golang/controllers"
	"suberes_golang/database"
	"suberes_golang/helpers"
	"suberes_golang/internal/appenv"
	"suberes_golang/internal/sentryutil"
	"suberes_golang/logger"
	middleware "suberes_golang/middlewares"
	"suberes_golang/queue"
	"suberes_golang/realtime"
	"suberes_golang/repositories"
	"suberes_golang/routes"
	"suberes_golang/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load .env file (no-op in production where vars are injected by Docker/systemd)
	godotenv.Load()

	// 2. Validate required env vars — panic early with a clear message if any are missing
	appenv.Validate()

	// 3. Structured logger (zerolog) — must be before any log.* call
	logger.Init()

	// 4. Sentry error tracking
	sentryutil.Init()

	config.ConnectDB()
	config.InitFirebase()
	database.AutoMigrate()

	realtime.InitSocket()

	// Use gin.New() instead of gin.Default() so we control Recovery & Logger middleware
	r := gin.New()

	r.SetTrustedProxies(nil)

	// ── Core middleware stack (order matters) ─────────────────────────────────
	r.Use(middleware.RecoveryMiddleware())           // panic → 500 + sentry report
	r.Use(middleware.RequestIDMiddleware())          // inject X-Request-ID
	r.Use(middleware.LoggerMiddleware())             // structured HTTP access log
	r.Use(middleware.TimeoutMiddleware(30 * time.Second)) // context deadline
	r.Use(middleware.SecurityHeadersMiddleware())   // OWASP security headers
	r.Use(middleware.RequestSizeLimiter(2 << 20))   // 2 MB request size cap

	// ── Probe endpoints (no auth, no rate-limit) ──────────────────────────────
	healthCtrl := &controllers.HealthController{}
	r.GET("/health", healthCtrl.Health)
	r.GET("/live", healthCtrl.Liveness)
	r.GET("/ready", healthCtrl.Readiness)

	// ── Root endpoint ─────────────────────────────────────────────────────────
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "suberes-api",
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// Serve static files for images (banners, etc)
	r.Static("/api/images", "./images")

	r.GET("/socket.io/*any", gin.WrapH(realtime.Server))
	r.POST("/socket.io/*any", gin.WrapH(realtime.Server))

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS") // comma-separated: "https://dashboard.suberes.com,https://app.suberes.com"
	var origins []string
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	} else {
		origins = []string{"http://localhost:5173", "http://localhost:3000"}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "device_language"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// i18n: baca header "device_language" dan simpan ke context sebelum handler manapun.
	r.Use(middleware.I18nMiddleware())

	userRepo := &repositories.UserRepository{DB: config.DB}
	userOtpRepo := &repositories.UserOtpRepository{DB: config.DB}
	bannerRepo := &repositories.BannerRepository{DB: config.DB}
	layananServiceRepo := &repositories.LayananServiceRepository{DB: config.DB}
	serviceRepo := &repositories.ServiceRepository{DB: config.DB}
	orderTransactionRepo := &repositories.OrderTransactionRepository{DB: config.DB}
	orderTransactionRepeatsRepo := &repositories.OrderTransactionRepeatsRepository{DB: config.DB}
	adminRepo := &repositories.AdminRepository{DB: config.DB}
	mitraRepo := &repositories.MitraRepository{DB: config.DB}
	subServiceRepo := &repositories.SubServiceRepository{DB: config.DB}
	subServiceAdditionalRepo := &repositories.SubServiceAdditionalRepository{DB: config.DB}
	paymentRepo := &repositories.PaymentRepository{DB: config.DB}
	subPaymentRepo := &repositories.SubPaymentRepository{DB: config.DB}
	subServiceAddedRepo := &repositories.SubServiceAddedRepository{DB: config.DB}
	orderRepo := &repositories.OrderRepository{DB: config.DB}
	orderChatRepo := &repositories.OrderChatRepository{DB: config.DB}
	orderOfferRepo := &repositories.OrderOfferRepository{DB: config.DB}
	bantuanRepo := &repositories.BantuanRepository{DB: config.DB}
	transactionRepo := &repositories.TransactionRepository{DB: config.DB}
	scheduleRepo := &repositories.ScheduleRepository{DB: config.DB}
	newsRepo := &repositories.NewsRepository{DB: config.DB}
	termsConditionRepo := &repositories.TermsConditionRepository{DB: config.DB}
	panduanRepo := &repositories.PanduanRepository{DB: config.DB}
	pinRepo := &repositories.PinRepository{DB: config.DB}
	ratingRepo := &repositories.RatingRepository{DB: config.DB}
	complainRepo := &repositories.ComplainRepository{DB: config.DB}

	customerService := &services.CustomerService{
		UserRepo:    userRepo,
		UserOTPRepo: userOtpRepo,
		DB:          config.DB,
	}

	bannerService := &services.BannerService{
		BannerRepo: bannerRepo,
		DB:         config.DB,
	}

	categoryServiceRepo := &repositories.CategoryServiceRepository{DB: config.DB}

	layananServiceService := &services.LayananServiceService{
		LayananServiceRepo:  layananServiceRepo,
		CategoryServiceRepo: categoryServiceRepo,
		DB:                  config.DB,
	}

	categoryServiceService := &services.CategoryServiceService{
		CategoryServiceRepo: categoryServiceRepo,
		UserRepo:            userRepo,
		DB:                  config.DB,
	}

	serviceService := &services.ServiceService{
		ServiceRepo: serviceRepo,
		DB:          config.DB,
	}

	adminService := &services.AdminService{
		OrderRepo: orderTransactionRepo,
		AdminRepo: adminRepo,
		DB:        config.DB,
	}

	mitraService := &services.MitraService{
		MitraRepository:                   mitraRepo,
		UserRepository:                    userRepo,
		UserOtpRepository:                 userOtpRepo,
		OrderTransactionRepository:        orderTransactionRepo,
		OrderTransactionRepeatsRepository: orderTransactionRepeatsRepo,
		ScheduleRepository:                scheduleRepo,
		OrderOfferRepository:              orderOfferRepo,
		DB:                                config.DB,
	}

	orderCashService := &services.OrderCashService{
		DB:                         config.DB,
		UserRepo:                   userRepo,
		ServiceRepo:                serviceRepo,
		SubServiceRepo:             subServiceRepo,
		LayananServiceRepo:         layananServiceRepo,
		SubPaymentRepo:             subPaymentRepo,
		SubServiceAddedRepo:        subServiceAddedRepo,
		PaymentRepo:                paymentRepo,
		OrderRepo:                  orderRepo,
		OrderChatRepo:              orderChatRepo,
		OrderOfferRepo:             orderOfferRepo,
		OrderTransactionRepo:       orderTransactionRepo,
		OrderTransactionRepeatRepo: orderTransactionRepeatsRepo,
	}

	subServiceService := &services.SubServiceService{
		SubServiceRepo:           subServiceRepo,
		SubServiceAdditionalRepo: subServiceAdditionalRepo,
		UserRepo:                 userRepo,
		DB:                       config.DB,
	}

	subServiceAdditionalService := &services.SubServiceAdditionalService{
		SubServiceAdditionalRepo: subServiceAdditionalRepo,
		UserRepo:                 userRepo,
		DB:                       config.DB,
	}

	subServiceController := &controllers.SubServiceController{
		SubServiceService: subServiceService,
	}

	subServiceAdditionalController := &controllers.SubServiceAdditionalController{
		SubServiceAdditionalService: subServiceAdditionalService,
	}

	bantuanService := &services.BantuanService{
		BantuanRepo: bantuanRepo,
		DB:          config.DB,
	}

	newsService := &services.NewsService{
		NewsRepo: newsRepo,
		DB:       config.DB,
	}

	newsController := &controllers.NewsController{
		NewsService: newsService,
	}

	termsConditionService := &services.TermsConditionService{
		TermsConditionRepo: termsConditionRepo,
		DB:                 config.DB,
	}

	termsConditionController := &controllers.TermsConditionController{
		TermsConditionService: termsConditionService,
	}

	panduanService := &services.PanduanService{
		PanduanRepo: panduanRepo,
		DB:          config.DB,
	}

	panduanController := &controllers.PanduanController{
		PanduanService: panduanService,
	}

	pinService := &services.PinService{
		PinRepo:     pinRepo,
		UserOTPRepo: userOtpRepo,
		DB:          config.DB,
	}
	pinController := &controllers.PinController{
		PinService: pinService,
	}

	ratingService := &services.RatingService{
		RatingRepo:           ratingRepo,
		OrderTransactionRepo: orderTransactionRepo,
		UserRepo:             userRepo,
		ServiceRepo:          serviceRepo,
		DB:                   config.DB,
	}
	ratingController := &controllers.RatingController{
		RatingService: ratingService,
	}

	complainService := &services.ComplainService{
		ComplainRepo: complainRepo,
		DB:           config.DB,
	}
	complainController := &controllers.ComplainController{
		ComplainService: complainService,
	}

	paymentService := &services.PaymentService{
		PaymentRepo: paymentRepo,
		DB:          config.DB,
	}

	paymentController := &controllers.PaymentController{
		PaymentService: paymentService,
	}

	subPaymentService := &services.SubPaymentService{
		DB: config.DB,
	}
	subPaymentController := &controllers.SubPaymentController{
		SubPaymentService: subPaymentService,
	}

	orderTransactionService := &services.OrderTransactionService{
		DB:                          config.DB,
		OrderTransactionRepo:        orderTransactionRepo,
		OrderTransactionRepeatsRepo: orderTransactionRepeatsRepo,
		OrderRepo:                   orderRepo,
		UserRepo:                    userRepo,
		SubServiceRepo:              subServiceRepo,
		TransactionRepo:             transactionRepo,
		PaymentRepo:                 paymentRepo,
		SubPaymentRepo:              subPaymentRepo,
		ServiceRepo:                 serviceRepo,
	}

	orderTransactionController := &controllers.OrderTransactionController{
		OrderTransactionService: orderTransactionService,
	}

	orderHistoryService := &services.OrderHistoryService{
		DB:                          config.DB,
		OrderTransactionRepo:        orderTransactionRepo,
		OrderTransactionRepeatsRepo: orderTransactionRepeatsRepo,
	}

	orderHistoryController := &controllers.OrderHistoryController{
		OrderHistoryService: orderHistoryService,
	}

	transactionService := services.NewTransactionService(transactionRepo)

	scheduleService := services.NewScheduleService(scheduleRepo, userRepo, config.DB)

	CustomerController := &controllers.CustomerController{
		CustomerService: customerService,
	}

	BannerController := &controllers.BannerController{
		BannerService: bannerService,
	}

	LayananServiceController := &controllers.LayananServiceController{
		LayananServiceService: layananServiceService,
	}

	CategoryServiceController := &controllers.CategoryServiceController{
		CategoryServiceService: categoryServiceService,
	}

	ServiceController := &controllers.ServiceController{
		ServiceService: serviceService,
	}

	AdminController := &controllers.AdminController{
		AdminService: adminService,
	}

	MitraController := &controllers.MitraController{
		MitraService: mitraService,
	}

	OrderController := controllers.NewOrderController(orderCashService)

	orderOfferService := services.NewOrderOfferService(config.DB)
	orderEwalletService := services.NewOrderEwalletService(config.DB)
	orderVAService := services.NewOrderVAService(config.DB)

	OrderOfferController := controllers.NewOrderOfferController(orderOfferService)
	OrderEwalletController := controllers.NewOrderEwalletController(orderEwalletService)
	OrderVAController := controllers.NewOrderVAController(orderVAService)

	BantuanController := &controllers.BantuanController{
		BantuanService: bantuanService,
	}

	TransactionController := controllers.NewTransactionController(transactionService)

	ScheduleController := controllers.NewScheduleController(scheduleService)

	api := r.Group("/api")
	api.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "suberes-api",
			"status":  "ok",
			"version": "1.0.0",
		})
	})
	routes.CustomerRoutes(api, CustomerController, config.DB)
	routes.BannerRoutes(api, BannerController, config.DB)
	routes.LayananServiceRoutes(api, LayananServiceController, config.DB)
	routes.CategoryServiceRoutes(api, CategoryServiceController, config.DB)
	routes.ServiceRoutes(api, ServiceController, config.DB)
	routes.SubServiceRoutes(api, subServiceController, config.DB)
	routes.SubServiceAdditionalRoutes(api, subServiceAdditionalController, config.DB)
	routes.AdminRoutes(api, AdminController, config.DB)
	routes.MitraRoutes(api, MitraController, config.DB)
	routes.OrderRoutes(api, OrderController, config.DB)
	routes.BantuanRoutes(api, BantuanController, config.DB)
	routes.TransactionRoutes(api, TransactionController, config.DB)
	routes.ScheduleRoutes(api, ScheduleController, config.DB)
	routes.NewsRoutes(api, newsController, config.DB)
	routes.TermsConditionRoutes(api, termsConditionController, config.DB)
	routes.PanduanRoutes(api, panduanController, config.DB)
	routes.PaymentRoutes(api, paymentController, config.DB)
	routes.SubPaymentRoutes(api, subPaymentController, config.DB)
	routes.OrderTransactionRoutes(api, orderTransactionController, config.DB)
	routes.OrderPendapatanRoutes(api, orderTransactionController, config.DB)
	routes.OrderOfferRoutes(api, OrderOfferController, config.DB)
	routes.OrderEwalletRoutes(api, OrderEwalletController, config.DB)
	routes.OrderVARoutes(api, OrderVAController, config.DB)

	webhookService := services.NewWebhookService(config.DB)
	webhookController := controllers.NewWebhookController(webhookService)
	routes.WebhookRoutes(api, webhookController)
	routes.OrderHistoryRoutes(api, orderHistoryController, config.DB)
	routes.PinRoutes(api, pinController, config.DB)
	routes.RatingRoutes(api, ratingController, config.DB)
	routes.ComplainRoutes(api, complainController, config.DB)

	disbursementService := services.NewDisbursementService(config.DB)
	disbursementController := controllers.NewDisbursementController(disbursementService)
	routes.DisbursementRoutes(api, disbursementController, config.DB)

	bankListService := services.NewBankListService(config.DB)
	bankListController := controllers.NewBankListController(bankListService)
	routes.BankListRoutes(api, bankListController, config.DB)

	queue.InitAsynq()

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	helpers.InitEmailQueue(redisHost + ":" + redisPort)

	go queue.StartWorker()

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Logger.Info().Str("port", port).Msg("server starting")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal().Err(err).Msg("server listen failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Logger.Info().Msg("shutdown signal received — draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	queue.StopWorker()
	sentryutil.Flush() // flush pending sentry events before exit

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Fatal().Err(err).Msg("graceful shutdown failed")
	}

	logger.Logger.Info().Msg("server exited cleanly")
}

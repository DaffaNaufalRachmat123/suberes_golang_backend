package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"suberes_golang/config"
	"suberes_golang/controllers"
	"suberes_golang/database"
	"suberes_golang/helpers"
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
	godotenv.Load()
	config.ConnectDB()
	config.InitFirebase()
	database.AutoMigrate()

	realtime.InitSocket()

	r := gin.Default()

	r.SetTrustedProxies(nil)

	r.GET("/socket.io/*any", gin.WrapH(realtime.Server))
	r.POST("/socket.io/*any", gin.WrapH(realtime.Server))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // atau isi domain frontend kamu
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

	subServiceController := &controllers.SubServiceController{
		SubServiceService: subServiceService,
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

	orderTransactionService := &services.OrderTransactionService{
		DB:                          config.DB,
		OrderTransactionRepo:        orderTransactionRepo,
		OrderTransactionRepeatsRepo: orderTransactionRepeatsRepo,
		OrderRepo:                   orderRepo,
		UserRepo:                    userRepo,
		SubServiceRepo:              subServiceRepo,
		TransactionRepo:             transactionRepo,
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
	routes.CustomerRoutes(api, CustomerController, config.DB)
	routes.BannerRoutes(api, BannerController, config.DB)
	routes.LayananServiceRoutes(api, LayananServiceController, config.DB)
	routes.ServiceRoutes(api, ServiceController, config.DB)
	routes.SubServiceRoutes(api, subServiceController, config.DB)
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
	routes.OrderTransactionRoutes(api, orderTransactionController, config.DB)
	routes.OrderOfferRoutes(api, OrderOfferController, config.DB)
	routes.OrderEwalletRoutes(api, OrderEwalletController, config.DB)
	routes.OrderVARoutes(api, OrderVAController, config.DB)
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

	log.Println("Server running on port", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	queue.StopWorker()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

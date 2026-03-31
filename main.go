package main

import (
	"log"
	"os"
	"suberes_golang/config"
	"suberes_golang/controllers"
	"suberes_golang/database"
	"suberes_golang/queue"
	"suberes_golang/realtime"
	"suberes_golang/repositories"
	"suberes_golang/routes"
	"suberes_golang/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.ConnectDB()
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

	orderService := services.NewOrderService(orderTransactionRepo)

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

	OrderController := controllers.NewOrderController(orderCashService, orderService)

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
	routes.AdminRoutes(api, AdminController, config.DB)
	routes.MitraRoutes(api, MitraController, config.DB)
	routes.OrderRoutes(api, OrderController, config.DB)
	routes.BantuanRoutes(api, BantuanController, config.DB)
	routes.TransactionRoutes(api, TransactionController, config.DB)
	routes.ScheduleRoutes(api, ScheduleController, config.DB)
	routes.NewsRoutes(api, newsController, config.DB)
	routes.TermsConditionRoutes(api, termsConditionController, config.DB)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)

	queue.InitAsynq()

	queue.StartWorker()
}

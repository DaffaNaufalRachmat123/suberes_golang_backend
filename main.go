package main

import (
	"log"
	"os"
	"suberes_golang/config"
	"suberes_golang/controllers"
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

	realtime.InitSocket()

	r := gin.Default()

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

	customerService := &services.CustomerService{
		UserRepo:    userRepo,
		UserOTPRepo: userOtpRepo,
		DB:          config.DB,
	}

	bannerService := &services.BannerService{
		BannerRepo: bannerRepo,
		DB:         config.DB,
	}

	layananServiceService := &services.LayananServiceService{
		LayananServiceRepo: layananServiceRepo,
		DB:                 config.DB,
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

	api := r.Group("/api")
	{
		routes.CustomerRoutes(api, CustomerController, config.DB)
		routes.BannerRoutes(api, BannerController, config.DB)
		routes.LayananServiceRoutes(api, LayananServiceController, config.DB)
		routes.ServiceRoutes(api, ServiceController, config.DB)
		routes.AdminRoutes(api, AdminController, config.DB)
		routes.MitraRoutes(api, MitraController, config.DB)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)
}

package main

import (
	"log"
	"os"
	"suberes_golang/config"
	"suberes_golang/controllers"
	"suberes_golang/repositories"
	"suberes_golang/routes"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	config.ConnectDB()

	r := gin.Default()

	userRepo := &repositories.UserRepository{DB: config.DB}
	userOtpRepo := &repositories.UserOtpRepository{DB: config.DB}
	bannerRepo := &repositories.BannerRepository{DB: config.DB}
	layananServiceRepo := &repositories.LayananServiceRepository{DB: config.DB}

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

	CustomerController := &controllers.CustomerController{
		CustomerService: customerService,
	}

	BannerController := &controllers.BannerController{
		BannerService: bannerService,
	}

	LayananServiceController := &controllers.LayananServiceController{
		LayananServiceService: layananServiceService,
	}

	api := r.Group("/api")
	{
		routes.CustomerRoutes(api, CustomerController, config.DB)
		routes.BannerRoutes(api, BannerController, config.DB)
		routes.LayananServiceRoutes(api, LayananServiceController, config.DB)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)
}

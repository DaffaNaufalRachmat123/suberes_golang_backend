package config

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func BuildDSN() string {
	env := strings.ToUpper(strings.TrimSpace(os.Getenv("APP_ENV")))

	var prefix string
	switch env {
	case "STAG":
		prefix = "STAG_"
	case "PROD":
		prefix = "PROD_"
	default:
		prefix = "DEV_"
	}

	host := os.Getenv(prefix + "HOST")
	user := os.Getenv(prefix + "USERNAME")
	pass := os.Getenv(prefix + "PASSWORD")
	dbname := os.Getenv(prefix + "DATABASE")
	port := os.Getenv(prefix + "PORT")

	if port == "" {
		port = "5432"
	}

	sslmode := os.Getenv(prefix + "SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		pass,
		host,
		port,
		dbname,
		sslmode,
	)
}

func ConnectDB() {
	dsn := BuildDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect DB: " + err.Error())
	}

	DB = db
}

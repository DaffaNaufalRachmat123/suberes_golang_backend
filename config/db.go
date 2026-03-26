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

	// Debug print
	fmt.Println("ENV:", env)
	fmt.Println("HOST:", host)
	fmt.Println("USER:", user)
	fmt.Println("DB:", dbname)
	fmt.Println("PORT:", port)

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user,
		pass,
		host,
		port,
		dbname,
	)
}

func ConnectDB() {
	dsn := BuildDSN()

	fmt.Println("DSN FULL:", dsn) // penting buat debug

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect DB: " + err.Error())
	}

	DB = db
}

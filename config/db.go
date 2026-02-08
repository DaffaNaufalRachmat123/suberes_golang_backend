package config

import (
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getEnv(key string) string {
	env := strings.ToUpper(strings.TrimSpace(os.Getenv("APP_ENV")))

	switch env {
	case "STAG":
		return os.Getenv("STAG_" + key)
	case "PROD":
		return os.Getenv("PROD_" + key)
	default: // DEV
		return os.Getenv("DEV_" + key)
	}
}

func ConnectDB() {

	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		getEnv("USERNAME"),
		getEnv("PASSWORD"),
		getEnv("HOST"),
		getEnv("DATABASE"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect DB: " + err.Error())
	}

	DB = db
}

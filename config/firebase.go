package config

import (
	"context"
	"log"
	"os"
	"strings"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var (
	CustomerApp *firebase.App
	MitraApp    *firebase.App
	AdminApp    *firebase.App
)

// firebaseCredDir returns the credential folder based on APP_ENV:
//   - prod  → credential_files_prod/
//   - dev / stag / (default) → credential_files/
//
// The folder can also be overridden explicitly via FIREBASE_CREDENTIAL_DIR env var.
func firebaseCredDir() string {
	if dir := strings.TrimSpace(os.Getenv("FIREBASE_CREDENTIAL_DIR")); dir != "" {
		return strings.TrimRight(dir, "/")
	}
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "prod" {
		return "credential_files_prod"
	}
	return "credential_files"
}

func InitFirebase() {
	ctx := context.Background()
	credDir := firebaseCredDir()

	var customerFile, mitraFile, adminFile string

	if strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) == "prod" {
		customerFile = credDir + "/suberes-prod-firebase-adminsdk-fbsvc-fd7a322244.json"
		mitraFile = credDir + "/suberes-mitra-prod-firebase-adminsdk-fbsvc-425edc434f.json"
		adminFile = credDir + "/suberes-dashboard-prod-firebase-adminsdk-fbsvc-fcd783ba9d.json"
	} else {
		customerFile = credDir + "/suberes-8b773-firebase-adminsdk-ci9gh-1db69d4c51.json"
		mitraFile = credDir + "/suberes-mitra-firebase-adminsdk-vml8u-072f804693.json"
		adminFile = credDir + "/suberes-dashboard-firebase-adminsdk-ns1j6-5218ef4faa.json"
	}

	customerOpt := option.WithCredentialsFile(customerFile)
	mitraOpt := option.WithCredentialsFile(mitraFile)
	adminOpt := option.WithCredentialsFile(adminFile)

	var err error

	CustomerApp, err = firebase.NewApp(ctx, nil, customerOpt)
	if err != nil {
		log.Fatalf("error initializing customer firebase: %v", err)
	}

	MitraApp, err = firebase.NewApp(ctx, nil, mitraOpt)
	if err != nil {
		log.Fatalf("error initializing mitra firebase: %v", err)
	}

	AdminApp, err = firebase.NewApp(ctx, nil, adminOpt)
	if err != nil {
		log.Fatalf("error initializing admin firebase: %v", err)
	}
}

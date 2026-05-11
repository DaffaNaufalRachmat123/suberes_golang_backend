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

	customerOpt := option.WithCredentialsFile(credDir + "/suberes-8b773-firebase-adminsdk-ci9gh-86aa7134d7.json")
	mitraOpt := option.WithCredentialsFile(credDir + "/suberes-mitra-firebase-adminsdk-vml8u-0977b1d80d.json")
	adminOpt := option.WithCredentialsFile(credDir + "/suberes-dashboard-firebase-adminsdk-ns1j6-5218ef4faa.json")

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

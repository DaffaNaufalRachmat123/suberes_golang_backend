package config

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var (
	CustomerApp *firebase.App
	MitraApp    *firebase.App
	AdminApp    *firebase.App
)

func InitFirebase() {
	ctx := context.Background()

	// Karena credential ada satu tingkat direktori
	customerOpt := option.WithCredentialsFile("../credential_files/suberes-8b773-firebase-adminsdk-ci9gh-647d90c6f7.json")
	mitraOpt := option.WithCredentialsFile("../credential_files/suberes-mitra-firebase-adminsdk-vml8u-b66d157f36.json")
	adminOpt := option.WithCredentialsFile("../credential_files/suberes-dashboard-firebase-adminsdk-ns1j6-5218ef4faa.json")

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

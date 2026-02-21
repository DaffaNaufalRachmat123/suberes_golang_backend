package service

import (
	"context"
	"fmt"
	"suberes_golang/config"
	"suberes_golang/helpers"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func SendTest(firebaseToken string) (string, error) {
	ctx := context.Background()

	client, err := config.MitraApp.Messaging(ctx)
	if err != nil {
		return "", err
	}

	msg := &messaging.Message{
		Token: firebaseToken,
	}

	return client.Send(ctx, msg)
}

func SendToDevice(db *gorm.DB, userType string, token string, payload map[string]interface{}) (string, error) {

	ctx := context.Background()

	payloadNotification := map[string]interface{}{
		"id":        uuid.New().String(),
		"user_type": userType,
	}

	var app *firebase.App

	switch userType {
	case "customer":
		app = config.CustomerApp
	case "mitra":
		app = config.MitraApp
	case "admin":
		app = config.AdminApp
	default:
		return "", fmt.Errorf("invalid user type")
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return "", err
	}

	tx := db.Begin()
	if tx.Error != nil {
		return "", tx.Error
	}

	msg := &messaging.Message{
		Token: token,
		Data:  payload["data"].(map[string]string),
	}

	response, err := client.Send(ctx, msg)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Copy optional fields
	data := payload["data"].(map[string]interface{})
	helpers.GetOtpDuration()
	helpers.CopyFields(data, payloadNotification)

	if err := tx.Table("notifications").Create(payloadNotification).Error; err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()
	return response, nil
}
func SendMulticast(db *gorm.DB, userType string, payload map[string]interface{}) (*messaging.BatchResponse, error) {

	ctx := context.Background()

	payloadNotification := map[string]interface{}{
		"id":        uuid.New().String(),
		"user_type": userType,
	}

	var app *firebase.App

	switch userType {
	case "customer":
		app = config.CustomerApp
	case "mitra":
		app = config.MitraApp
	case "admin":
		app = config.AdminApp
	default:
		return nil, fmt.Errorf("invalid user type")
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	msg := &messaging.MulticastMessage{
		Tokens: payload["tokens"].([]string),
		Data:   payload["data"].(map[string]string),
	}

	response, err := client.SendMulticast(ctx, msg)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	data := payload["data"].(map[string]interface{})
	helpers.CopyFields(data, payloadNotification)

	if err := tx.Table("notifications").Create(payloadNotification).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return response, nil
}

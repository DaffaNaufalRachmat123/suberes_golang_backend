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

	var strData map[string]string
	var dataForCopy map[string]interface{}
	switch d := payload["data"].(type) {
	case map[string]string:
		strData = d
		dataForCopy = make(map[string]interface{}, len(d))
		for k, v := range d {
			dataForCopy[k] = v
		}
	case map[string]interface{}:
		dataForCopy = d
		strData = make(map[string]string, len(d))
		for k, v := range d {
			strData[k] = fmt.Sprintf("%v", v)
		}
	default:
		tx.Rollback()
		return "", fmt.Errorf("invalid data type in payload")
	}

	msg := &messaging.Message{
		Token: token,
		Data:  strData,
	}

	response, err := client.Send(ctx, msg)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	// Copy optional fields
	helpers.GetOtpDuration()
	helpers.CopyFields(dataForCopy, payloadNotification)

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

	var strData map[string]string
	var rawData map[string]interface{}
	switch d := payload["data"].(type) {
	case map[string]string:
		strData = d
		rawData = make(map[string]interface{}, len(d))
		for k, v := range d {
			rawData[k] = v
		}
	case map[string]interface{}:
		rawData = d
		strData = make(map[string]string, len(d))
		for k, v := range d {
			strData[k] = fmt.Sprintf("%v", v)
		}
	default:
		tx.Rollback()
		return nil, fmt.Errorf("invalid data type in payload")
	}

	msg := &messaging.MulticastMessage{
		Tokens: payload["tokens"].([]string),
		Data:   strData,
	}

	responses, err := client.SendEachForMulticast(ctx, msg)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	helpers.CopyFields(rawData, payloadNotification)

	if err := tx.Table("notifications").Create(payloadNotification).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	// Adapt to BatchResponse-like return for compatibility
	return responses, nil
}

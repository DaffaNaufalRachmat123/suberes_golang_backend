package controllers

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"suberes_golang/config"
	"suberes_golang/dtos"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type XenditController struct {
	service *services.XenditService
}

func NewXenditController() *XenditController {
	return &XenditController{
		service: services.NewXenditService(config.DB),
	}
}

// EwalletCallback handles all e-wallet related callbacks from Xendit.
func (ctrl *XenditController) EwalletCallback(c *gin.Context) {
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body

	var payload dtos.XenditCallbackPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse JSON"})
		return
	}
	log.Printf("Xendit Callback Received: Event: %s", payload.Event)

	var serviceErr error
	switch payload.Event {
	case "ewallet.capture":
		serviceErr = ctrl.service.HandleEwalletCapture(&payload.Data)
	case "ewallet.void":
		serviceErr = ctrl.service.HandleEwalletVoid(&payload.Data)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event type not handled"})
		return
	}

	if serviceErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (ctrl *XenditController) CallbackNotification(c *gin.Context) {
	var body interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	log.Printf("Generic Callback Received: %+v", body)
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

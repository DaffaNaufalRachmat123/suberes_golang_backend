package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	PaymentService *services.PaymentService
}

// GET /payments/index
// Returns all active payments with their enabled sub_payments (no pagination –
// payment methods are a small, finite list).
func (c *PaymentController) Index(ctx *gin.Context) {
	data, err := c.PaymentService.GetAllActive()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

// POST /payments/create  (multipart: json_data + file)
func (c *PaymentController) Create(ctx *gin.Context) {
	if err := c.PaymentService.Create(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "payment created",
		"status":         "success",
	})
}

// PUT /payments/update/image/:id  (multipart: json_data + file)
func (c *PaymentController) UpdateWithImage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "invalid id",
			"status":         "failure",
		})
		return
	}
	if err := c.PaymentService.UpdateWithImage(ctx, id); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "Payment not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "payment updated",
		"status":         "success",
	})
}

// PUT /payments/update/:id  (JSON body)
func (c *PaymentController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "invalid id",
			"status":         "failure",
		})
		return
	}
	if err := c.PaymentService.Update(ctx, id); err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "Payment not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "payment updated",
		"status":         "success",
	})
}

// DELETE /payments/remove/:id
func (c *PaymentController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "invalid id",
			"status":         "failure",
		})
		return
	}
	if err := c.PaymentService.Delete(id); err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "Payment not found" {
			statusCode = http.StatusNotFound
		}
		ctx.JSON(statusCode, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "payment deleted",
		"status":         "success",
	})
}

package controllers

import (
	"net/http"
	"suberes_golang/dtos"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type SubServiceAdditionalController struct {
	SubServiceAdditionalService *services.SubServiceAdditionalService
}

func (c *SubServiceAdditionalController) Create(ctx *gin.Context) {
	var req dtos.CreateSubServiceAdditionalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "Invalid data request", "status": "failure", "error": err.Error()})
		return
	}

	additional, err := c.SubServiceAdditionalService.Create(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "sub service additional created",
		"status":         "success",
		"data":           additional,
	})
}

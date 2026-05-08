package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/models"
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

func (c *SubServiceAdditionalController) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "Invalid ID", "status": "failure"})
		return
	}

	var req dtos.UpdateSubServiceAdditionalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "Invalid data request", "status": "failure", "error": err.Error()})
		return
	}

	req.ID = id

	updated, err := c.SubServiceAdditionalService.Update(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "sub service additional updated",
		"status":         "success",
		"data":           updated,
	})
}

func (c *SubServiceAdditionalController) Delete(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "Invalid ID", "status": "failure"})
		return
	}

	var req dtos.DeleteSubServiceAdditionalRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "Invalid data request", "status": "failure", "error": err.Error()})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": "unauthorized", "status": "failure"})
		return
	}

	user, ok := userCtx.(models.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": "unauthorized", "status": "failure"})
		return
	}

	if err := c.SubServiceAdditionalService.Delete(id, user.ID, req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "sub service additional removed",
		"status":         "success",
	})
}

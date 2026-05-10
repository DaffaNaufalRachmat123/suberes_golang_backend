package controllers

import (
	"suberes_golang/i18n"
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/models"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type SubServiceController struct {
	SubServiceService *services.SubServiceService
}

func (c *SubServiceController) Create(ctx *gin.Context) {
	var req dtos.SubServiceCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidPayload), "status": "failure", "error": err.Error()})
		return
	}

	subService, err := c.SubServiceService.Create(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgSubServiceCreated), "status": "success", "data": subService})
}

func (c *SubServiceController) Update(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidID), "status": "failure"})
		return
	}

	var req dtos.SubServiceUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidPayload), "status": "failure", "error": err.Error()})
		return
	}

	req.ID = id

	updated, err := c.SubServiceService.Update(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgSubServiceUpdated), "status": "success", "data": updated})
}

func (c *SubServiceController) Delete(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidID), "status": "failure"})
		return
	}

	var req dtos.SubServiceDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidPayload), "status": "failure", "error": err.Error()})
		return
	}

	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgUnauthorized), "status": "failure"})
		return
	}

	user, ok := userCtx.(models.User)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgUnauthorized), "status": "failure"})
		return
	}

	if err := c.SubServiceService.Delete(id, user.ID, req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgSubServiceRemoved), "status": "success"})
}

package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/i18n"
	"suberes_golang/models"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type CategoryServiceController struct {
	CategoryServiceService *services.CategoryServiceService
}

func (c *CategoryServiceController) GetDetail(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidID), "status": "failure"})
		return
	}

	data, err := c.CategoryServiceService.GetDetail(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgOK), "status": "success", "data": data})
}

func (c *CategoryServiceController) Create(ctx *gin.Context) {
	var req dtos.CategoryServiceRequest
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

	if err := c.CategoryServiceService.Create(req.LayananID, req.CategoryService, user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgCategoryCreated), "status": "success"})
}

func (c *CategoryServiceController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidID), "status": "failure"})
		return
	}

	var req dtos.CategoryServiceRequest
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

	if err := c.CategoryServiceService.Update(uint(id), req.LayananID, req.CategoryService, user.ID); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgCategoryUpdated), "status": "success"})
}

func (c *CategoryServiceController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgInvalidID), "status": "failure"})
		return
	}

	var req dtos.CategoryServiceDeleteRequest
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

	if err := c.CategoryServiceService.Delete(uint(id), user.ID, req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": i18n.Tc(ctx, i18n.MsgCategoryDeleted), "status": "success"})
}

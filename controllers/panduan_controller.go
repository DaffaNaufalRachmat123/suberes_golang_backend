package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type PanduanController struct {
	PanduanService *services.PanduanService
}

func (c *PanduanController) IndexCustomer(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	panduans, total, err := c.PanduanService.GetPanduansCustomer(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}

	response := helpers.GetPaginationData(ctx, panduans, len(panduans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *PanduanController) IndexMitra(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	panduans, total, err := c.PanduanService.GetPanduansMitra(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}

	response := helpers.GetPaginationData(ctx, panduans, len(panduans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *PanduanController) IndexAdmin(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	panduans, total, err := c.PanduanService.GetPanduansAdmin(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}

	response := helpers.GetPaginationData(ctx, panduans, len(panduans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *PanduanController) Detail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Invalid ID",
			"status":         "failure",
		})
		return
	}

	panduan, err := c.PanduanService.GetPanduanByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}

	if panduan == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"server_message": "Panduan not found",
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, panduan)
}

func (c *PanduanController) Create(ctx *gin.Context) {
	if err := c.PanduanService.CreatePanduan(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Bad request",
			"status":         "failed",
			"error":          err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Panduan created",
		"status":         "success",
	})
}

func (c *PanduanController) UpdateWatchingCount(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Invalid ID",
			"status":         "failure",
		})
		return
	}

	if err := c.PanduanService.UpdateWatchingCount(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Watching count updated",
		"status":         "success",
	})
}

func (c *PanduanController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Invalid ID",
			"status":         "failure",
		})
		return
	}

	if err := c.PanduanService.UpdatePanduan(ctx, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Bad request",
			"status":         "failed",
			"error":          err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Panduan updated",
		"status":         "success",
	})
}

func (c *PanduanController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "Invalid ID",
			"status":         "failure",
		})
		return
	}

	if err := c.PanduanService.DeletePanduan(uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failed",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Panduan removed",
		"status":         "success",
	})
}

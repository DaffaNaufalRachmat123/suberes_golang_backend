package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type BantuanController struct {
	BantuanService *services.BantuanService
}

func (c *BantuanController) IndexCustomer(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	bantuans, total, err := c.BantuanService.GetBantuans(page, limit, "customer")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, bantuans, len(bantuans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *BantuanController) IndexMitra(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	bantuans, total, err := c.BantuanService.GetBantuans(page, limit, "mitra")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, bantuans, len(bantuans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *BantuanController) IndexAdmin(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	bantuans, total, err := c.BantuanService.GetBantuansAdmin(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, bantuans, len(bantuans), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *BantuanController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	data, err := c.BantuanService.GetBantuanByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *BantuanController) Create(ctx *gin.Context) {
	if err := c.BantuanService.CreateBantuan(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "bantuan created",
		"status":         "success",
	})
}

func (c *BantuanController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.BantuanService.UpdateBantuan(ctx, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "bantuan updated",
		"status":         "success",
	})
}

func (c *BantuanController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.BantuanService.DeleteBantuan(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "bantuan removed",
		"status":         "success",
	})
}

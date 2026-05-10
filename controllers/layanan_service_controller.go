package controllers

import (
	"suberes_golang/i18n"
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type LayananServiceController struct {
	LayananServiceService *services.LayananServiceService
}

func (c *LayananServiceController) Index(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	layananServices, total, err := c.LayananServiceService.GetLayananService(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInternalError),
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, layananServices, len(layananServices), 1, 5, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *LayananServiceController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	data, err := c.LayananServiceService.GetLayananByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInternalError),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *LayananServiceController) GetPopular(ctx *gin.Context) {
	data, err := c.LayananServiceService.GetLayananPopular()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInternalError),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *LayananServiceController) Create(ctx *gin.Context) {
	if err := c.LayananServiceService.Create(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgLayananCreated),
		"status":         "success",
	})
}

// PUT /api/banner/update/image/:id
func (c *LayananServiceController) Update(ctx *gin.Context) {
	if err := c.LayananServiceService.Update(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgBannerUpdated),
		"status":         "success",
	})
}

// DELETE /api/banner/remove/:id
func (c *LayananServiceController) GetDetail(ctx *gin.Context) {
	layananID, err := strconv.Atoi(ctx.Param("layanan_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInvalidLayananID),
			"status":         "failure",
		})
		return
	}

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	categoryServices, total, err := c.LayananServiceService.GetCategoryServiceByLayananID(layananID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInternalError),
			"status":         "failure",
		})
		return
	}

	response := helpers.GetPaginationData(ctx, categoryServices, len(categoryServices), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *LayananServiceController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.LayananServiceService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgBannerRemoved),
		"status":         "success",
	})
}

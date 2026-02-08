package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type BannerController struct {
	BannerService *services.BannerService
}

func (c *BannerController) Index(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	banners, total, err := c.BannerService.GetBanners(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, banners, len(banners), 1, 5, total)
	ctx.JSON(http.StatusOK, response)
}
func (c *BannerController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	data, err := c.BannerService.GetBannerByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (c *BannerController) GetPopular(ctx *gin.Context) {
	data, err := c.BannerService.GetPopularBanners()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (c *BannerController) Create(ctx *gin.Context) {
	if err := c.BannerService.CreateBanner(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "banner created",
		"status":         "success",
	})
}

// PUT /api/banner/update/image/:id
func (c *BannerController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.BannerService.UpdateBanner(ctx, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "banner updated",
		"status":         "success",
	})
}

// DELETE /api/banner/remove/:id
func (c *BannerController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.BannerService.DeleteBanner(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "banner removed",
		"status":         "success",
	})
}

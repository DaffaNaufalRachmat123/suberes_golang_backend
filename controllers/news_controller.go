package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type NewsController struct {
	NewsService *services.NewsService
}

func (c *NewsController) Index(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	news, total, err := c.NewsService.GetNews(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	response := helpers.GetPaginationData(ctx, news, len(news), 1, 5, total)
	ctx.JSON(http.StatusOK, response)
}
func (c *NewsController) GetByID(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	data, err := c.NewsService.GetNewsByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (c *NewsController) GetPopular(ctx *gin.Context) {
	data, err := c.NewsService.GetPopularNews()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, data)
}
func (c *NewsController) Create(ctx *gin.Context) {
	if err := c.NewsService.CreateNews(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "news created",
		"status":         "success",
	})
}

func (c *NewsController) Update(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.NewsService.UpdateNews(ctx, uint(id)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "news updated",
		"status":         "success",
	})
}

func (c *NewsController) Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if err := c.NewsService.DeleteNews(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "failure",
			"error":  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "news removed",
		"status":         "success",
	})
}

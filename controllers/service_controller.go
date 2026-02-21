package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	ServiceService *services.ServiceService
}

func (c *ServiceController) Index(ctx *gin.Context) {
	parent_id, err := strconv.Atoi(ctx.Param("parent_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "bad request",
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
	services, total, err := c.ServiceService.GetServices(parent_id, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": "Internal server error",
			"status":         "failure",
		})
	}
	response := helpers.GetPaginationData(ctx, services, len(services), 1, 5, total)
	ctx.JSON(http.StatusOK, response)
}
func (c *ServiceController) LayananServices(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("layanan_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "bad request",
			"status":         "failure",
		})
		return
	}
	layananServices, err := c.ServiceService.GetLayananServices(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, layananServices)
}
func (c *ServiceController) Popular(ctx *gin.Context) {
	servicePopular, err := c.ServiceService.GetPopular()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, servicePopular)
}

func (c *ServiceController) GetDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	service, err := c.ServiceService.GetServiceByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(http.StatusOK, service)
}
func (c *ServiceController) SearchService(ctx *gin.Context) {
	layananId, err := strconv.Atoi(ctx.Param("layanan_id"))
	serviceName := ctx.DefaultQuery("service_name", "")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	data, err := c.ServiceService.ServiceRepo.Search(layananId, serviceName)
	ctx.JSON(http.StatusOK, data)
}
func (c *ServiceController) ServiceType(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	data, err := c.ServiceService.ServiceRepo.FindServiceType(id)
	ctx.JSON(http.StatusOK, data)
}
func (c *ServiceController) Create(ctx *gin.Context) {
	if err := c.ServiceService.Create(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Service created",
		"status":         "success",
	})
}
func (c *ServiceController) UpdateImage(ctx *gin.Context) {
	if err := c.ServiceService.UpdateImage(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Service updated with image",
		"status":         "success",
	})
}

func (c *ServiceController) Update(ctx *gin.Context) {
	if err := c.ServiceService.Update(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Service updated",
		"status":         "success",
	})
}
func (c *ServiceController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	var req dtos.ServiceDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}
	userCtx, exists := ctx.Get("currentUser")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Customer not found",
		})
	}
	user := userCtx.(models.User)
	if err := c.ServiceService.Delete(id, user.ID, req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Service deleted",
		"status":         "success",
	})
}

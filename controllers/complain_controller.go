package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ComplainController struct {
	ComplainService *services.ComplainService
}

// IndexAdmin GET /api/complains/index  (admin, search by complain_code)
func (c *ComplainController) IndexAdmin(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	search := ctx.DefaultQuery("search", "")

	data, total, err := c.ComplainService.GetAllAdmin(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	response := helpers.GetPaginationData(ctx, data, len(data), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// IndexCustomer GET /api/complains/index/customer
func (c *ComplainController) IndexCustomer(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	data, total, err := c.ComplainService.GetAllCustomer(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	response := helpers.GetPaginationData(ctx, data, len(data), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// IndexMitra GET /api/complains/index/mitra
func (c *ComplainController) IndexMitra(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	data, total, err := c.ComplainService.GetAllMitra(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	response := helpers.GetPaginationData(ctx, data, len(data), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// Detail GET /api/complains/detail/:id
func (c *ComplainController) Detail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure"})
		return
	}
	data, err := c.ComplainService.GetDetail(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "complain not found", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

// Create POST /api/complains/create  (multipart/form-data with optional file[] images)
func (c *ComplainController) Create(ctx *gin.Context) {
	if err := c.ComplainService.Create(ctx); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "complain created", "status": "success"})
}

// UpdateStatus PUT /api/complains/update/:id/:status
func (c *ComplainController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure"})
		return
	}
	status := ctx.Param("status")
	if err := c.ComplainService.UpdateStatus(id, status); err != nil {
		if err.Error() == "complain not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "complain not found", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "complain updated", "status": "success"})
}

// Remove DELETE /api/complains/remove/:id
func (c *ComplainController) Remove(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "bad request", "status": "failure"})
		return
	}
	if err := c.ComplainService.Remove(id); err != nil {
		if err.Error() == "complain not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": "complain not found", "status": "failure"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"server_message": "complain removed", "status": "success"})
}

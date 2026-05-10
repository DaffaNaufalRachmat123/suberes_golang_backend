package controllers

import (
	"suberes_golang/i18n"
	"net/http"
	"strconv"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type SubPaymentController struct {
	SubPaymentService *services.SubPaymentService
}

// GET /sub_payments/index?page=1&limit=10
func (c *SubPaymentController) Index(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	data, total, err := c.SubPaymentService.GetAll(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInternalError),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GET /sub_payments/detail/:id
func (c *SubPaymentController) Detail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInvalidID),
			"status":         "failure",
		})
		return
	}

	data, err := c.SubPaymentService.GetByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   data,
	})
}

// PUT /sub_payments/update/:id
func (c *SubPaymentController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": i18n.Tc(ctx, i18n.MsgInvalidID),
			"status":         "failure",
		})
		return
	}

	data, err := c.SubPaymentService.Update(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": i18n.Tc(ctx, i18n.MsgSubPaymentUpdated),
		"status":         "success",
		"data":           data,
	})
}

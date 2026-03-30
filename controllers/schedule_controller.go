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

type ScheduleController struct {
	ScheduleService *services.ScheduleService
}

func NewScheduleController(scheduleService *services.ScheduleService) *ScheduleController {
	return &ScheduleController{
		ScheduleService: scheduleService,
	}
}

func (c *ScheduleController) Index(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.DefaultQuery("search", "")

	schedules, total, err := c.ScheduleService.Index(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	pagination := helpers.GetPaginationData(ctx, schedules, len(schedules), page, limit, total)

	ctx.JSON(http.StatusOK, pagination)
}

func (c *ScheduleController) IndexMitra(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.DefaultQuery("search", "")

	schedules, total, err := c.ScheduleService.IndexMitra(page, limit, search)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	pagination := helpers.GetPaginationData(ctx, schedules, len(schedules), page, limit, total)

	ctx.JSON(http.StatusOK, pagination)
}

func (c *ScheduleController) Detail(ctx *gin.Context) {
	id := ctx.Param("id")
	schedule, err := c.ScheduleService.Detail(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, schedule)
}

func (c *ScheduleController) Create(ctx *gin.Context) {
	var req dtos.CreateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := c.ScheduleService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"server_message": "schedule created",
		"status":         "success",
		"data":           schedule,
	})
}

func (c *ScheduleController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dtos.UpdateScheduleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule, err := c.ScheduleService.Update(id, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "schedule updated",
		"status":         "success",
		"data":           schedule,
	})
}

func (c *ScheduleController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	password := ctx.Param("password")

	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)

	err := c.ScheduleService.Delete(id, password, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "schedule removed",
		"status":         "success",
	})
}

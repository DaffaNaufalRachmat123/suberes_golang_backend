package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type RatingController struct {
	RatingService *services.RatingService
}

// GetMitraRatings GET /api/ratings/mitra/:id/:limit/:offset
func (c *RatingController) GetMitraRatings(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	limit, err := strconv.Atoi(ctx.Param("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset, err := strconv.Atoi(ctx.Param("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	ratings, err := c.RatingService.GetMitraRatings(mitraID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, ratings)
}

// GetCustomerRatings GET /api/ratings/customer/:id/:limit/:offset
func (c *RatingController) GetCustomerRatings(ctx *gin.Context) {
	mitraID := ctx.Param("id")
	limit, err := strconv.Atoi(ctx.Param("limit"))
	if err != nil || limit < 1 {
		limit = 10
	}
	offset, err := strconv.Atoi(ctx.Param("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	ratings, err := c.RatingService.GetCustomerRatings(mitraID, limit, offset)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, ratings)
}

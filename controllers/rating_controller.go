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

// GetMitraRatingHome GET /api/ratings/mitra/home/:mitra_id
func (c *RatingController) GetMitraRatingHome(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")

	response, err := c.RatingService.GetMitraRatingHome(mitraID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// GetMitraRatingsPaginated GET /api/ratings/mitra/list/:mitra_id?page=1&limit=10
func (c *RatingController) GetMitraRatingsPaginated(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")

	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	ratings, total, err := c.RatingService.GetMitraRatingsPaginated(mitraID, page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	ctx.JSON(http.StatusOK, helpers.GetPaginationData(ctx, ratings, len(ratings), page, limit, total))
}

// CreateRatingToMitra POST /api/ratings/create_to_mitra/:order_id/:customer_id/:mitra_id/:rating
func (c *RatingController) CreateRatingToMitra(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	ratingVal, err := strconv.ParseFloat(ctx.Param("rating"), 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid rating value", "status": "failure"})
		return
	}

	var req services.CreateRatingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	if err := c.RatingService.CreateRatingToMitra(orderID, customerID, mitraID, ratingVal, req); err != nil {
		switch err.Error() {
		case "order not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
		case "customer or mitra data not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
		default:
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "rating submitted", "status": "success"})
}

// CreateRatingToCustomer POST /api/ratings/create_to_customer/:order_id/:customer_id/:mitra_id/:rating
func (c *RatingController) CreateRatingToCustomer(ctx *gin.Context) {
	orderID := ctx.Param("order_id")
	customerID := ctx.Param("customer_id")
	mitraID := ctx.Param("mitra_id")

	ratingVal, err := strconv.ParseFloat(ctx.Param("rating"), 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": "invalid rating value", "status": "failure"})
		return
	}

	var req services.CreateRatingRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"server_message": err.Error(), "status": "failure"})
		return
	}

	if err := c.RatingService.CreateRatingToCustomer(orderID, customerID, mitraID, ratingVal, req); err != nil {
		switch err.Error() {
		case "order not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
		case "customer or mitra data not found":
			ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
		default:
			if errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"server_message": err.Error(), "status": "failure"})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"server_message": err.Error(), "status": "failure"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"server_message": "rating submitted", "status": "success"})
}

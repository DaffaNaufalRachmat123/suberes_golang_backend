package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type BankListController struct {
	BankListService *services.BankListService
}

func NewBankListController(bankListService *services.BankListService) *BankListController {
	return &BankListController{
		BankListService: bankListService,
	}
}

// GetTopupBanks returns paginated banks eligible for topup.
// GET /bank_list/topup
func (c *BankListController) GetTopupBanks(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	banks, total, err := c.BankListService.GetTopupBanks(page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, banks, len(banks), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// GetDisbursementBanks returns paginated banks eligible for disbursement.
// GET /bank_list
func (c *BankListController) GetDisbursementBanks(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	banks, total, err := c.BankListService.GetDisbursementBanks(page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, banks, len(banks), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// BulkCreateBanks bulk-inserts banks with method_type = 'bank'.
// POST /bank_list/:admin_id
func (c *BankListController) BulkCreateBanks(ctx *gin.Context) {
	adminID := ctx.Param("admin_id")

	var items []dtos.BankListCreateItem
	if err := ctx.ShouldBindJSON(&items); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "bad request",
			"status":         "failure",
			"error":          err.Error(),
		})
		return
	}

	if err := c.BankListService.BulkCreateBanks(adminID, items); err != nil {
		if err.Error() == "admin not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"server_message": "admin not found",
				"status":         "failure",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Bank List Created",
		"status":         "success",
	})
}

// BulkCreateEwallets bulk-inserts banks with method_type = 'ewallet'.
// POST /ewallet_list/:admin_id
func (c *BankListController) BulkCreateEwallets(ctx *gin.Context) {
	adminID := ctx.Param("admin_id")

	var items []dtos.BankListCreateItem
	if err := ctx.ShouldBindJSON(&items); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "bad request",
			"status":         "failure",
			"error":          err.Error(),
		})
		return
	}

	if err := c.BankListService.BulkCreateEwallets(adminID, items); err != nil {
		if err.Error() == "admin not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"server_message": "admin not found",
				"status":         "failure",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Ewallet List Created",
		"status":         "success",
	})
}

// UpdateBankEwallet updates a bank/ewallet entry by id.
// PUT /bank_ewallet_list/:id
func (c *BankListController) UpdateBankEwallet(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "invalid id",
			"status":         "failure",
		})
		return
	}

	var req dtos.BankListUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"server_message": "bad request",
			"status":         "failure",
			"error":          err.Error(),
		})
		return
	}

	if err := c.BankListService.Update(id, &req); err != nil {
		if err.Error() == "bank not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
				"server_message": "bank not found",
				"status":         "failure",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Bank/Ewallet updated",
		"status":         "success",
	})
}

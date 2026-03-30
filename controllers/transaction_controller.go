package controllers

import (
	"net/http"
	"strconv"
	"suberes_golang/helpers"
	"suberes_golang/services"

	"github.com/gin-gonic/gin"
)

type TransactionController struct {
	TransactionService *services.TransactionService
}

func NewTransactionController(transactionService *services.TransactionService) *TransactionController {
	return &TransactionController{
		TransactionService: transactionService,
	}
}

func (c *TransactionController) FindAll(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	search := ctx.DefaultQuery("search", "")
	transactionType := ctx.DefaultQuery("transaction_type", "")

	transactions, total, err := c.TransactionService.FindAllWithPagination(page, limit, search, transactionType)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, transactions, len(transactions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *TransactionController) GetTransactionTypes(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	date := ctx.Param("pendapatan_date")

	types, err := c.TransactionService.GetTransactionTypesByMitraIDAndDate(mitraID, date)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transaction types", http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, types)
}

func (c *TransactionController) FindAllByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	transactionFor := ctx.Param("transaction_for")
	transactionTime := ctx.Param("transaction_time")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	transactions, total, err := c.TransactionService.FindAllByMitraIDWithPagination(mitraID, transactionFor, transactionTime, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, transactions, len(transactions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

func (c *TransactionController) FindDisbursementsByMitra(ctx *gin.Context) {
	mitraID := ctx.Param("mitra_id")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	transactions, total, err := c.TransactionService.FindDisbursementsByMitraID(mitraID, page, limit)
	if err != nil {
		helpers.APIErrorResponse(ctx, "Failed to get transactions", http.StatusInternalServerError)
		return
	}

	response := helpers.GetPaginationData(ctx, transactions, len(transactions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

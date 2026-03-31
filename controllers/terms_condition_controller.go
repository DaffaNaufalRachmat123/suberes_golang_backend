package controllers

import (
"net/http"
"strconv"
"suberes_golang/dtos"
"suberes_golang/helpers"
"suberes_golang/services"

"github.com/gin-gonic/gin"
)

type TermsConditionController struct {
	TermsConditionService *services.TermsConditionService
}

// Index retrieves all terms conditions with pagination (admin only)
func (c *TermsConditionController) Index(ctx *gin.Context) {
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}

	termsConditions, total, err := c.TermsConditionService.GetAllTermsConditions(page, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Internal server error",
"status":         "failure",
})
		return
	}

	response := helpers.GetPaginationData(ctx, termsConditions, len(termsConditions), page, limit, total)
	ctx.JSON(http.StatusOK, response)
}

// Detail retrieves a single terms condition by ID (admin only)
func (c *TermsConditionController) Detail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Invalid ID",
"status":         "failure",
})
		return
	}

	termsCondition, err := c.TermsConditionService.GetTermsConditionByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Internal server error",
"status":         "failure",
})
		return
	}

	if termsCondition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
"server_message": "Terms condition not found",
"status":         "failure",
})
		return
	}

	ctx.JSON(http.StatusOK, termsCondition)
}

// GetByTypeAndUserType retrieves active terms condition for users (public endpoint)
func (c *TermsConditionController) GetByTypeAndUserType(ctx *gin.Context) {
	tocType := ctx.Param("toc_type")
	tocUserType := ctx.Param("toc_user_type")

	if tocType == "" || tocUserType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Missing required parameters",
"status":         "failure",
})
		return
	}

	termsCondition, err := c.TermsConditionService.GetTermsConditionByTypeAndUserType(tocType, tocUserType)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Internal server error",
"status":         "failure",
})
		return
	}

	if termsCondition == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
"server_message": "Terms condition not found",
"status":         "failure",
})
		return
	}

	ctx.JSON(http.StatusOK, termsCondition)
}

// Create creates a new terms condition (admin only)
func (c *TermsConditionController) Create(ctx *gin.Context) {
	forceParam := ctx.Param("force")
	force := forceParam == "true"

	var req dtos.TermsConditionCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Bad request",
"status":         "failure",
"error":          err.Error(),
		})
		return
	}

	// Get creator ID from JWT token
	creatorID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
"server_message": "Unauthorized",
"status":         "failure",
})
		return
	}

	if err := c.TermsConditionService.CreateTermsCondition(&req, creatorID.(string), force); err != nil {
		// Check if conflict error
		if err.Error() == "There is still an active TOC for this user" {
			ctx.JSON(http.StatusConflict, gin.H{
"server_message": err.Error(),
				"status":         "failed",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Failed to create terms condition",
"status":         "failure",
"error":          err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
"server_message": "toc created",
"status":         "succeeded",
})
}

// UpdateStatus updates the status of a terms condition (admin only)
func (c *TermsConditionController) UpdateStatus(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Invalid ID",
"status":         "failure",
})
		return
	}

	var req dtos.TermsConditionUpdateStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Bad Request",
"status":         "failure",
"error":          err.Error(),
		})
		return
	}

	if err := c.TermsConditionService.UpdateTermsConditionStatus(uint(id), &req); err != nil {
		// Check specific error messages
		errorMsg := err.Error()
		if errorMsg == "Masih ada TOC yang aktif untuk sasaran ini" {
			ctx.JSON(http.StatusForbidden, gin.H{
"server_message": errorMsg,
"status":         "failure",
})
			return
		}
		if errorMsg == "TOC not found" {
			ctx.JSON(http.StatusNotFound, gin.H{
"server_message": errorMsg,
"status":         "failed",
})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Failed to update status",
"status":         "failure",
"error":          errorMsg,
})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
"server_message": "terms condition updated",
"status":         "success",
})
}

// Update updates a terms condition (admin only)
func (c *TermsConditionController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Invalid ID",
"status":         "failure",
})
		return
	}

	forceParam := ctx.Param("force")
	force := forceParam == "true"

	var req dtos.TermsConditionUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Bad request",
"status":         "failure",
"error":          err.Error(),
		})
		return
	}

	// Get creator ID from JWT token
	creatorID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
"server_message": "Unauthorized",
"status":         "failure",
})
		return
	}

	if err := c.TermsConditionService.UpdateTermsCondition(uint(id), &req, creatorID.(string), force); err != nil {
		errorMsg := err.Error()

		// Check specific error messages
		if errorMsg == "There are no TOC data" {
			ctx.JSON(http.StatusNotFound, gin.H{
"server_message": errorMsg,
"status":         "failed",
})
			return
		}
		if errorMsg == "Masih ada TOC aktif dengan kategori untuk user ini" {
			ctx.JSON(http.StatusConflict, gin.H{
"server_message": errorMsg,
"status":         "failed",
})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Failed to update terms condition",
"status":         "failure",
"error":          errorMsg,
})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
"server_message": "toc updated",
"status":         "succeeded",
})
}

// Delete deletes a terms condition (admin only)
func (c *TermsConditionController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
"server_message": "Invalid ID",
"status":         "failure",
})
		return
	}

	if err := c.TermsConditionService.DeleteTermsCondition(uint(id)); err != nil {
		errorMsg := err.Error()

		if errorMsg == "Data TOC tak ditemukan" {
			ctx.JSON(http.StatusNotFound, gin.H{
"server_message": errorMsg,
"status":         "failure",
})
			return
		}
		if errorMsg == "Harap nonaktifkan TOC dahulu" {
			ctx.JSON(http.StatusForbidden, gin.H{
"server_message": errorMsg,
"status":         "failure",
})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
"server_message": "Failed to delete terms condition",
"status":         "failure",
"error":          errorMsg,
})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
"server_message": "Berhasil menghapus TOC",
"status":         "success",
})
}

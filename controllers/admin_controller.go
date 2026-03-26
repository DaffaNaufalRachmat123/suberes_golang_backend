package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"suberes_golang/dtos"
	"suberes_golang/helpers"
	"suberes_golang/models"
	"suberes_golang/services"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	AdminService *services.AdminService
}

func (c *AdminController) GetDashboard(ctx *gin.Context) {
	dashboardPayload, err := c.AdminService.GetDashboard()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}
	ctx.JSON(http.StatusOK, dashboardPayload)
}

func (c *AdminController) IndexAdmin(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)

	admins, total, err := c.AdminService.IndexAdmin(page, limit, user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  admins,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (c *AdminController) CreateAdmin(ctx *gin.Context) {
	var req dtos.CreateAdminRequest
	jsonData := ctx.Request.FormValue("json_data")
	if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	pathString := filepath.Join(
		helpers.RootPath(),
		os.Getenv("IMAGE_PATH_CONTROLLER"),
	)
	if _, err := os.Stat(pathString); os.IsNotExist(err) {
		os.Mkdir(pathString, os.ModePerm)
	}

	adminImagePath := pathString + os.Getenv("ADMIN_IMAGE_PATH")
	if _, err := os.Stat(adminImagePath); os.IsNotExist(err) {
		os.Mkdir(adminImagePath, os.ModePerm)
	}

	filePath := fmt.Sprintf("ADMIN_IMG_%d_%s", time.Now().UnixNano(), file.Filename)
	if err := ctx.SaveUploadedFile(file, adminImagePath+filePath); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	dbFilePath := os.Getenv("ADMIN_IMAGE_PATH") + filePath

	admin, err := c.AdminService.CreateAdmin(&req, dbFilePath)
	fmt.Printf("%+v\n", admin)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"server_message": "admin created",
		"status":         "success",
		"callback":       admin,
	})
}

func (c *AdminController) UpdateAdminStatus(ctx *gin.Context) {
	adminID := ctx.Param("admin_id")
	var req dtos.UpdateAdminStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.AdminService.UpdateAdminStatus(adminID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failed",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "admin status updated",
		"status":         "success",
	})
}

func (c *AdminController) RemoveAdmin(ctx *gin.Context) {
	adminID := ctx.Param("admin_id")
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)

	var req dtos.RemoveAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.AdminService.RemoveAdmin(adminID, user.ID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failed",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "remove account succeed",
		"status":         "success",
	})
}

func (c *AdminController) Login(ctx *gin.Context) {
	var req dtos.LoginAdminRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, user, err := c.AdminService.Login(&req)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "Berhasil masuk. Selamat datang kembali",
		"status":         "success",
		"token":          "Bearer " + token,
		"data":           user,
	})
}

func (c *AdminController) Logout(ctx *gin.Context) {
	userCtx, _ := ctx.Get("currentUser")
	user := userCtx.(models.User)

	if err := c.AdminService.Logout(user.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "logout successful",
		"status":         "ok",
		"isLogout":       true,
	})
}

func (c *AdminController) UpdateFirebaseToken(ctx *gin.Context) {
	userID := ctx.Param("id")
	var req dtos.UpdateFirebaseTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.AdminService.UpdateFirebaseToken(userID, req.FirebaseToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"server_message": err.Error(),
			"status":         "failure",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"server_message": "refresh token updated",
		"status":         "success",
	})
}

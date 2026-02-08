package helpers

import "github.com/gin-gonic/gin"

// BaseResponse adalah struktur universal JSON kamu
type BaseResponse struct {
	ServerMessage string      `json:"server_message"`
	Status        string      `json:"status"`
	Data          interface{} `json:"data"` // interface{} artinya bisa diisi struct apa saja / null
}

// 1. Helper untuk Response SUKSES
func APIResponse(ctx *gin.Context, message string, status int, data interface{}) {
	jsonResponse := BaseResponse{
		ServerMessage: message,
		Status:        "OK", // Atau bisa dinamis sesuai parameter
		Data:          data,
	}

	if status >= 400 {
		jsonResponse.Status = "failure"
	}

	ctx.JSON(status, jsonResponse)
}

// 2. Helper untuk Response ERROR
func APIErrorResponse(ctx *gin.Context, message string, status int) {
	jsonResponse := BaseResponse{
		ServerMessage: message,
		Status:        "failure",
		Data:          nil, // Data kosong kalau error
	}
	ctx.JSON(status, jsonResponse)
}

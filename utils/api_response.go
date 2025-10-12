package utils

import "github.com/gin-gonic/gin"

// APIResponse adalah struktur standar untuk response JSON.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SuccessResponse membuat format response standar untuk sukses.
// Menggunakan gin.H agar lebih fleksibel dan langsung kompatibel dengan c.JSON.
func SuccessResponse(message string, data interface{}) gin.H {
	return gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	}
}

// ErrorResponse membuat format response standar untuk error.
func ErrorResponse(message string, errorDetails interface{}) gin.H {
	// Jika tidak ada detail error, kita set datanya ke null
	if errorDetails == nil {
		errorDetails = nil
	}
	
	return gin.H{
		"status":  "error",
		"message": message,
		"data":    errorDetails,
	}
}
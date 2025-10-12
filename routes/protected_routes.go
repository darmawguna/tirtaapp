package routes

import (
	"net/http"

	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin" // Ganti path module Anda
)

func SetupProtectedRoutes(router *gin.Engine) {
	// Buat grup baru untuk API v1
	apiV1 := router.Group("/api/v1")
	
	// Terapkan middleware autentikasi ke seluruh grup ini
	apiV1.Use(middlewares.AuthMiddleware())
	{
		// Endpoint untuk mengetes
		apiV1.GET("/profile", func(c *gin.Context) {
			// Ambil data user dari context yang sudah di-set oleh middleware
			userID := c.MustGet("userID")
			userRole := c.MustGet("userRole")

			c.JSON(http.StatusOK, gin.H{
				"message":  "Welcome to the protected area!",
				"user_id":  userID,
				"user_role": userRole,
			})
		})
	}
}
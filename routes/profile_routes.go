package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers" // Sesuaikan path
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

// SetupProfileRoutes mendaftarkan endpoint untuk profil pengguna.
func SetupProfileRoutes(router *gin.Engine, handler *handlers.ProfileHandler) {
	// Grup /api/v1 sudah seharusnya ada dan menggunakan AuthMiddleware
	// Jika belum, Anda perlu membuatnya di main.go atau di sini.
	// Kita asumsikan grup /api/v1 sudah ada.
	profileRoutes := router.Group("/api/v1")
	profileRoutes.Use(middlewares.AuthMiddleware()) // Pastikan middleware terpasang
	{
		profileRoutes.GET("/profile", handler.GetProfile)
		profileRoutes.PUT("/profile", handler.UpdateProfile)
	}

	adminRoutes := router.Group("/api/v1/admin")
	adminRoutes.Use(middlewares.AuthMiddleware())   // Pertama, pastikan login
	adminRoutes.Use(middlewares.AdminMiddleware()) // Kedua, pastikan adalah admin
	{
		adminRoutes.GET("/users/count", handler.GetUserCount)
	}
}
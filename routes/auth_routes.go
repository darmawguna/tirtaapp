package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes mendaftarkan semua route yang berhubungan dengan autentikasi.
func SetupAuthRoutes(router *gin.Engine, authHandler *handlers.AuthHandler) {
	// Grouping route untuk /auth
	authRoutes := router.Group("/api/v1/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		authRoutes.POST("/login", authHandler.Login)
		// Nanti route /login akan kita tambahkan di sini
	}
}
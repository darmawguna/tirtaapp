package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupComplaintRoutes(router *gin.Engine, complaintHandler *handlers.ComplaintHandler) {
	// Grup untuk endpoint yang memerlukan autentikasi
	complaintRoutes := router.Group("/api/v1/complaints")
	complaintRoutes.Use(middlewares.AuthMiddleware())
	{
		// Endpoint untuk membuat keluhan baru
		complaintRoutes.POST("/", complaintHandler.Create)
		// Endpoint untuk melihat riwayat keluhan sendiri
		complaintRoutes.GET("/", complaintHandler.GetMyComplaints) // <-- TAMBAHKAN INI
		complaintRoutes.GET("/:id", complaintHandler.GetDetailComplaint)
	}
}
package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers" // Sesuaikan path jika berbeda
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

// SetupMedicationRefillRoutes mendaftarkan endpoint untuk jadwal obat habis.
func SetupMedicationRefillRoutes(router *gin.Engine, handler *handlers.MedicationRefillHandler) {
	// Buat grup route di bawah /api/v1 dan terapkan middleware otentikasi
	routes := router.Group("/api/v1/medication-refills")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.POST("/", handler.Create)
		routes.GET("/", handler.GetAll)
		routes.GET("/:id", handler.GetByID)
		routes.PUT("/:id", handler.Update)
		routes.DELETE("/:id", handler.Delete)
	}
}
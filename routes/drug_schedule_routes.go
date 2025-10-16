package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupDrugScheduleRoutes(router *gin.Engine, handler *handlers.DrugScheduleHandler) {
	// Semua route di sini memerlukan login
	scheduleRoutes := router.Group("/api/v1/drug-schedules")
	scheduleRoutes.Use(middlewares.AuthMiddleware())
	{
		scheduleRoutes.POST("/", handler.Create)
		scheduleRoutes.GET("/", handler.GetAll)
		scheduleRoutes.GET("/:id", handler.GetByID)
		scheduleRoutes.PUT("/:id", handler.Update)
		scheduleRoutes.DELETE("/:id", handler.Delete)
	}
}
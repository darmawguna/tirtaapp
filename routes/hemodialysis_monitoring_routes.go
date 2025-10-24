package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers" // Adjust path
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupHemodialysisMonitoringRoutes(router *gin.Engine, handler *handlers.HemodialysisMonitoringHandler) {
	routes := router.Group("/api/v1/hemodialysis-monitoring")
	routes.Use(middlewares.AuthMiddleware())
	{
		// Create monitoring data for a specific schedule
		routes.POST("/", handler.Create)
		// Get monitoring history for the logged-in user
		routes.GET("/history", handler.GetHistory)
		routes.GET("/:id", handler.GetByID)
		// Get specific monitoring data by schedule ID (Add handler if needed)
		// routes.GET("/:schedule_id", handler.GetByScheduleID)
	}
}
package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupHemodialysisScheduleRoutes(router *gin.Engine, handler *handlers.HemodialysisScheduleHandler) {
	routes := router.Group("/api/v1/hemodialysis-schedules")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.POST("/", handler.Create)
		routes.GET("/", handler.GetAll)
		routes.PUT("/:id", handler.Update)
		routes.GET("/:id", handler.GetByID)
		routes.DELETE("/:id", handler.Delete)
	}
}
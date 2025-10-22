package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers" // Sesuaikan path
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupFluidBalanceRoutes(router *gin.Engine, handler *handlers.FluidBalanceHandler) {
	routes := router.Group("/api/v1/fluids")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.POST("/", handler.CreateOrUpdate) // Endpoint untuk input harian
		routes.GET("/", handler.GetHistory)     // Endpoint untuk riwayat
	}
}
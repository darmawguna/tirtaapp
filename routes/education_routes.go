package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupEducationRoutes(router *gin.Engine, educationHandler *handlers.EducationHandler) {
	// Semua route edukasi memerlukan login.
	educationRoutes := router.Group("/api/v1/educations")
	educationRoutes.Use(middlewares.AuthMiddleware())
	{
		// Semua user yang login bisa melihat daftar edukasi.
		educationRoutes.GET("/", educationHandler.GetAll)
		educationRoutes.GET("/:id", educationHandler.GetByID)

		// Hanya Admin yang bisa CUD.
		adminRoutes := educationRoutes.Group("/")
		adminRoutes.Use(middlewares.AdminMiddleware())
		{
			adminRoutes.POST("/", educationHandler.Create)
			adminRoutes.PUT("/:id", educationHandler.Update)
			adminRoutes.DELETE("/:id", educationHandler.Delete)
		}
	}
}
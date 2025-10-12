package routes

import (
	"github.com/darmawguna/tirtaapp.git/handlers"
	middlewares "github.com/darmawguna/tirtaapp.git/middleware"
	"github.com/gin-gonic/gin"
)

func SetupQuizRoutes(router *gin.Engine, quizHandler *handlers.QuizHandler) {
	// Semua route kuis memerlukan login
	quizRoutes := router.Group("/api/v1/quizzes")

	quizRoutes.Use(middlewares.AuthMiddleware())
	{
		// Hanya user biasa & admin bisa melihat semua kuis
		quizRoutes.GET("/", quizHandler.GetAll)
		quizRoutes.GET("/:id", quizHandler.GetByID)
		// Hanya ADMIN yang bisa membuat, update, dan delete
		adminRoutes := quizRoutes.Group("/")
		adminRoutes.Use(middlewares.AdminMiddleware())
		{
			adminRoutes.POST("/", quizHandler.Create)
			adminRoutes.PUT("/:id", quizHandler.Update)
			adminRoutes.DELETE("/:id", quizHandler.Delete)
		}
	}
}

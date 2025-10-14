package main

import (
	"log"
	"os"

	"github.com/darmawguna/tirtaapp.git/config"
	"github.com/darmawguna/tirtaapp.git/handlers"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/routes"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	config.LoadConfig()
	db = config.ConnectDB() // Koneksi setelah config dimuat
	config.RunMigration(db)

	// Inisialisasi Gin router
	router := gin.Default()

	// Inisialisasi semua layer (Dependency Injection)
	userRepository := repositories.NewUserRepository(db)
	quizRepository := repositories.NewQuizRepository(db) 
	educationRepository := repositories.NewEducationRepository(db)
	complaintRepository := repositories.NewComplaintRepository(db)

	authService := services.NewAuthService(userRepository)
	quizService := services.NewQuizService(quizRepository)
	educationService := services.NewEducationService(educationRepository)
	complaintService := services.NewComplaintService(complaintRepository)

	authHandler := handlers.NewAuthHandler(authService)
	quizHandler := handlers.NewQuizHandler(quizService)
	educationHandler := handlers.NewEducationHandler(educationService)
	complaintHandler := handlers.NewComplaintHandler(complaintService)

	
	// Mendaftarkan routes dari file terpisah
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupProtectedRoutes(router)
	routes.SetupQuizRoutes(router, quizHandler)
	routes.SetupEducationRoutes(router, educationHandler)
	routes.SetupComplaintRoutes(router, complaintHandler)

	// Simple health check route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API is running"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server is running on port %s", port)
	router.Run(":" + port)
}

package main

import (
	"log"
	"os"

	"github.com/darmawguna/tirtaapp.git/config"
	"github.com/darmawguna/tirtaapp.git/handlers"
	models "github.com/darmawguna/tirtaapp.git/model"
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
	config.RunMigration(db,
		&models.User{},
		&models.Quiz{},
		&models.Education{},
		&models.ComplaintLog{},
		&models.DrugSchedule{},
		&models.ControlSchedule{},      // <-- Jangan lupa tambahkan model baru
		&models.HemodialysisSchedule{}, // <-- Jangan lupa tambahkan model baru
		&models.Device{},               // <-- Jangan lupa tambahkan model baru
	)
	queueService := services.NewQueueService()
	if err := queueService.Connect(); err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %s", err)
	}
	// Pastikan koneksi ditutup saat aplikasi berhenti
	defer queueService.Close()

	// Inisialisasi Gin router
	router := gin.Default()
	// Inisialisasi semua layer (Dependency Injection)
	userRepository := repositories.NewUserRepository(db)
	quizRepository := repositories.NewQuizRepository(db)
	educationRepository := repositories.NewEducationRepository(db)
	complaintRepository := repositories.NewComplaintRepository(db)
	drugScheduleRepository := repositories.NewDrugScheduleRepository(db)
	deviceRepository := repositories.NewDeviceRepository(db)
	controlScheduleRepo := repositories.NewControlScheduleRepository(db)
    hemodialysisScheduleRepo := repositories.NewHemodialysisScheduleRepository(db)

	deviceService := services.NewDeviceService(deviceRepository)
	authService := services.NewAuthService(userRepository, deviceService)
	quizService := services.NewQuizService(quizRepository)
	educationService := services.NewEducationService(educationRepository)
	complaintService := services.NewComplaintService(complaintRepository)
	drugScheduleService := services.NewDrugScheduleService(drugScheduleRepository, queueService)
	controlScheduleService := services.NewControlScheduleService(controlScheduleRepo, queueService)
    hemodialysisScheduleService := services.NewHemodialysisScheduleService(hemodialysisScheduleRepo, queueService)

	authHandler := handlers.NewAuthHandler(authService)
	quizHandler := handlers.NewQuizHandler(quizService)
	educationHandler := handlers.NewEducationHandler(educationService)
	complaintHandler := handlers.NewComplaintHandler(complaintService)
	drugScheduleHandler := handlers.NewDrugScheduleHandler(drugScheduleService)
	controlScheduleHandler := handlers.NewControlScheduleHandler(controlScheduleService)
    hemodialysisScheduleHandler := handlers.NewHemodialysisScheduleHandler(hemodialysisScheduleService)

	// Mendaftarkan routes dari file terpisah
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupProtectedRoutes(router)
	routes.SetupQuizRoutes(router, quizHandler)
	routes.SetupEducationRoutes(router, educationHandler)
	routes.SetupComplaintRoutes(router, complaintHandler)
	routes.SetupDrugScheduleRoutes(router, drugScheduleHandler)
	routes.SetupControlScheduleRoutes(router, controlScheduleHandler)
    routes.SetupHemodialysisScheduleRoutes(router, hemodialysisScheduleHandler)

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

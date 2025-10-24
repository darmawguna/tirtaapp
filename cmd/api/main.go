package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darmawguna/tirtaapp.git/utils"

	"github.com/darmawguna/tirtaapp.git/config"
	"github.com/darmawguna/tirtaapp.git/handlers"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/routes"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// --- Tahap 1: Inisialisasi Konfigurasi & Koneksi ---
	clearDB := flag.Bool("clear-db", false, "Set this flag to delete all data from the database before starting")
	flag.Parse()
	config.LoadConfig()
	db = config.ConnectDB()
	config.RunMigration(db,
		&models.User{}, &models.Quiz{}, &models.Education{}, &models.ComplaintLog{},
		&models.DrugSchedule{}, &models.ControlSchedule{}, &models.HemodialysisSchedule{},
		&models.Device{}, &models.FluidBalanceLog{}, &models.HemodialysisMonitoring{},
	)

	// Inisialisasi Queue Service (RabbitMQ)
	queueService := services.NewQueueService()
	if err := queueService.Connect(); err != nil {
		log.Fatalf("Could not connect to RabbitMQ: %s", err)
	}
	if *clearDB {
		err := utils.ClearAllData(db) // Call the clear function
		if err != nil {
			log.Fatalf("FATAL: Could not clear database: %v", err)
		}
		log.Println("Database cleared successfully via flag.")
		// Optional: os.Exit(0) if you only want to clear and not start the server
	}

	// --- Tahap 2: Dependency Injection ---
	// Inisialisasi semua layer (Repository, Service, Handler)
	userRepository := repositories.NewUserRepository(db)
	quizRepository := repositories.NewQuizRepository(db)
	educationRepository := repositories.NewEducationRepository(db)
	deviceRepository := repositories.NewDeviceRepository(db)
	drugScheduleRepository := repositories.NewDrugScheduleRepository(db)
	controlScheduleRepo := repositories.NewControlScheduleRepository(db)
	hemodialysisScheduleRepo := repositories.NewHemodialysisScheduleRepository(db)
	fluidBalanceRepo := repositories.NewFluidBalanceRepository(db)
	hemodialysisMonitoringRepo := repositories.NewHemodialysisMonitoringRepository(db)
	complaintRepository := repositories.NewComplaintRepository(db)
	// (Tambahkan repository lain di sini jika ada)

	deviceService := services.NewDeviceService(deviceRepository)
	quizService := services.NewQuizService(quizRepository)
	educationService := services.NewEducationService(educationRepository)
	authService := services.NewAuthService(userRepository, deviceService)
	drugScheduleService := services.NewDrugScheduleService(drugScheduleRepository, queueService)
	controlScheduleService := services.NewControlScheduleService(controlScheduleRepo, queueService)
	hemodialysisScheduleService := services.NewHemodialysisScheduleService(hemodialysisScheduleRepo, queueService)
	fluidBalanceService := services.NewFluidBalanceService(fluidBalanceRepo, userRepository)
	hemodialysisMonitoringService := services.NewHemodialysisMonitoringService(hemodialysisMonitoringRepo, userRepository)
	profileService := services.NewProfileService(userRepository)
	complaintService := services.NewComplaintService(complaintRepository)
	// (Tambahkan service lain di sini jika ada)

	authHandler := handlers.NewAuthHandler(authService)
	drugScheduleHandler := handlers.NewDrugScheduleHandler(drugScheduleService)
	quizHandler := handlers.NewQuizHandler(quizService)
	educationHandler := handlers.NewEducationHandler(educationService)
	controlScheduleHandler := handlers.NewControlScheduleHandler(controlScheduleService)
	hemodialysisScheduleHandler := handlers.NewHemodialysisScheduleHandler(hemodialysisScheduleService)
	fluidBalanceHandler := handlers.NewFluidBalanceHandler(fluidBalanceService)
	hemodialysisMonitoringHandler := handlers.NewHemodialysisMonitoringHandler(hemodialysisMonitoringService)
	profileHandler := handlers.NewProfileHandler(profileService)
	complaintHandler := handlers.NewComplaintHandler(complaintService)
	// (Tambahkan handler lain di sini jika ada)

	// --- Tahap 3: Setup Router dan Server ---
	router := gin.Default()
	router.Static("/static", "./uploads")

	// Mendaftarkan semua routes
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupDrugScheduleRoutes(router, drugScheduleHandler)
	routes.SetupQuizRoutes(router,quizHandler)
	routes.SetupEducationRoutes(router, educationHandler)
	routes.SetupControlScheduleRoutes(router, controlScheduleHandler)
	routes.SetupHemodialysisScheduleRoutes(router, hemodialysisScheduleHandler)
	routes.SetupFluidBalanceRoutes(router, fluidBalanceHandler)
	routes.SetupHemodialysisMonitoringRoutes(router,hemodialysisMonitoringHandler)
	routes.SetupProfileRoutes(router, profileHandler)
	routes.SetupComplaintRoutes(router, complaintHandler)

	// (Tambahkan pendaftaran route lain di sini)

	// Endpoint health check
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API is running"})
	})

	// Konfigurasi server HTTP
	srv := &http.Server{
		Addr:    ":" + viper.GetString("PORT"),
		Handler: router,
	}

	// --- Tahap 4: Jalankan Server & Graceful Shutdown ---
	// Jalankan server HTTP di sebuah goroutine agar tidak memblokir.
	go func() {
		log.Printf("Server is running on port %s", viper.GetString("PORT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Siapkan channel untuk mendengarkan sinyal shutdown dari sistem operasi
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Blokir eksekusi sampai sinyal shutdown diterima
	<-quit
	log.Println("Shutting down server...")

	// Tutup koneksi RabbitMQ dengan bersih
	queueService.Close()
	log.Println("RabbitMQ connection closed.")

	// Beri waktu 5 detik bagi server untuk menyelesaikan request yang sedang berjalan.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

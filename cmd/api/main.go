package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// --- Tahap 2: Dependency Injection ---
	// Inisialisasi semua layer (Repository, Service, Handler)
	userRepository := repositories.NewUserRepository(db)
	deviceRepository := repositories.NewDeviceRepository(db)
	drugScheduleRepository := repositories.NewDrugScheduleRepository(db)
	controlScheduleRepo := repositories.NewControlScheduleRepository(db)
	hemodialysisScheduleRepo := repositories.NewHemodialysisScheduleRepository(db)
	fluidBalanceRepo := repositories.NewFluidBalanceRepository(db)
	hemodialysisMonitoringRepo := repositories.NewHemodialysisMonitoringRepository(db)
	// (Tambahkan repository lain di sini jika ada)

	deviceService := services.NewDeviceService(deviceRepository)
	authService := services.NewAuthService(userRepository, deviceService)
	drugScheduleService := services.NewDrugScheduleService(drugScheduleRepository, queueService)
	controlScheduleService := services.NewControlScheduleService(controlScheduleRepo, queueService)
	hemodialysisScheduleService := services.NewHemodialysisScheduleService(hemodialysisScheduleRepo, queueService)
	fluidBalanceService := services.NewFluidBalanceService(fluidBalanceRepo, userRepository)
	hemodialysisMonitoringService := services.NewHemodialysisMonitoringService(hemodialysisMonitoringRepo, hemodialysisScheduleRepo)
	// (Tambahkan service lain di sini jika ada)

	authHandler := handlers.NewAuthHandler(authService)
	drugScheduleHandler := handlers.NewDrugScheduleHandler(drugScheduleService)
	controlScheduleHandler := handlers.NewControlScheduleHandler(controlScheduleService)
	hemodialysisScheduleHandler := handlers.NewHemodialysisScheduleHandler(hemodialysisScheduleService)
	fluidBalanceHandler := handlers.NewFluidBalanceHandler(fluidBalanceService)
	hemodialysisMonitoringHandler := handlers.NewHemodialysisMonitoringHandler(hemodialysisMonitoringService)
	// (Tambahkan handler lain di sini jika ada)

	// --- Tahap 3: Setup Router dan Server ---
	router := gin.Default()

	// Mendaftarkan semua routes
	routes.SetupAuthRoutes(router, authHandler)
	routes.SetupDrugScheduleRoutes(router, drugScheduleHandler)
	routes.SetupControlScheduleRoutes(router, controlScheduleHandler)
	routes.SetupHemodialysisScheduleRoutes(router, hemodialysisScheduleHandler)
	routes.SetupFluidBalanceRoutes(router, fluidBalanceHandler)
	routes.SetupHemodialysisMonitoringRoutes(router,hemodialysisMonitoringHandler)
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

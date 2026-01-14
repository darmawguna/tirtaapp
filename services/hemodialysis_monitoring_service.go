package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm"
)

// ✅ DEMO MODE: Hard-code tanggal untuk sidang
const DEMO_DATE_HEMODIALYSIS = "2026-01-14" // Ubah ini setelah sidang selesai

// Interface service
type HemodialysisMonitoringService interface {
	CreateOrUpdateMonitoringForToday(user models.User, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error)
	GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error)
	GetMonitoringByID(userID, monitoringID uint) (models.HemodialysisMonitoring, error)
}

// Implementasi service
type hemodialysisMonitoringService struct {
	monitoringRepo repositories.HemodialysisMonitoringRepository
	userRepo       repositories.UserRepository
}

// Constructor
func NewHemodialysisMonitoringService(monitoringRepo repositories.HemodialysisMonitoringRepository, userRepo repositories.UserRepository) HemodialysisMonitoringService {
	return &hemodialysisMonitoringService{
		monitoringRepo: monitoringRepo,
		userRepo:       userRepo,
	}
}

// CreateOrUpdateMonitoringForToday: Logika bisnis utama
func (s *hemodialysisMonitoringService) CreateOrUpdateMonitoringForToday(
	user models.User,
	input dto.CreateHemodialysisMonitoringDTO,
) (models.HemodialysisMonitoring, error) {

	// ===============================
	// 1. Validasi dan Load Timezone User
	// ===============================
	if user.Timezone == "" {
		return models.HemodialysisMonitoring{}, fmt.Errorf("user timezone not set")
	}

	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf(
			"invalid user timezone (%s): %w",
			user.Timezone,
			err,
		)
	}

	// ===============================
	// 2. ✅ GUNAKAN TANGGAL DEMO UNTUK SIDANG
	// ===============================
	// PRODUCTION: Uncomment code di bawah dan hapus DEMO_DATE_HEMODIALYSIS
	// nowUser := time.Now().In(loc)
	// todayUser := time.Date(
	// 	nowUser.Year(),
	// 	nowUser.Month(),
	// 	nowUser.Day(),
	// 	0, 0, 0, 0,
	// 	loc,
	// )

	// ✅ DEMO MODE: Paksa tanggal ke 14 Januari 2026
	todayUser, err := time.ParseInLocation("2006-01-02", DEMO_DATE_HEMODIALYSIS, loc)
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf("failed to parse demo date: %w", err)
	}

	// ===============================
	// 3. Validasi Input (Opsional tapi bagus untuk production)
	// ===============================
	if input.BPBefore == "" || input.BPAfter == "" {
		return models.HemodialysisMonitoring{}, fmt.Errorf("tekanan darah sebelum dan sesudah harus diisi")
	}
	if input.WeightBefore <= 0 || input.WeightAfter <= 0 {
		return models.HemodialysisMonitoring{}, fmt.Errorf("berat badan harus lebih dari 0")
	}

	// ===============================
	// 4. Cari Monitoring Hari Ini
	// ===============================
	existing, err := s.monitoringRepo.FindByUserIDAndDate(
		user.ID,
		todayUser, // ✅ Akan selalu 2026-01-14
	)

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.HemodialysisMonitoring{}, fmt.Errorf(
				"failed to check today's monitoring: %w",
				err,
			)
		}

		// ===============================
		// 5A. CREATE - Jika Belum Ada
		// ===============================
		newMonitoring := models.HemodialysisMonitoring{
			UserID:         user.ID,
			MonitoringDate: todayUser, // ✅ Akan selalu 2026-01-14
			BPBefore:       input.BPBefore,
			BPAfter:        input.BPAfter,
			WeightBefore:   input.WeightBefore,
			WeightAfter:    input.WeightAfter,
		}

		return s.monitoringRepo.Create(newMonitoring)
	}

	// ===============================
	// 5B. UPDATE - Jika Sudah Ada
	// ===============================
	existing.BPBefore = input.BPBefore
	existing.BPAfter = input.BPAfter
	existing.WeightBefore = input.WeightBefore
	existing.WeightAfter = input.WeightAfter

	return s.monitoringRepo.Update(existing)
}

// GetMonitoringHistory mengambil riwayat
func (s *hemodialysisMonitoringService) GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error) {
	// ✅ Validasi limit
	limit := 10
	if limit <= 0 || limit > 100 {
		limit = 10 // Default safe limit
	}

	// Ambil data terakhir
	history, err := s.monitoringRepo.FindHistoryByUserID(userID, limit)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat monitoring: %w", err)
	}
	return history, nil
}

func (s *hemodialysisMonitoringService) GetMonitoringByID(userID, monitoringID uint) (models.HemodialysisMonitoring, error) {
	monitoring, err := s.monitoringRepo.FindByID(monitoringID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.HemodialysisMonitoring{}, errors.New("data pemantauan tidak ditemukan")
		}
		return models.HemodialysisMonitoring{}, fmt.Errorf("gagal mencari data pemantauan: %w", err)
	}

	// ✅ Verifikasi kepemilikan untuk security
	if monitoring.UserID != userID {
		return models.HemodialysisMonitoring{}, errors.New("tidak berwenang mengakses data pemantauan ini")
	}

	return monitoring, nil
}
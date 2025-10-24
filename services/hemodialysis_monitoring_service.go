package services

import (
	"errors" // <-- Pastikan import ini ada
	"fmt"

	// "log" // Tidak perlu log di sini, error dikembalikan
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"          // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm" // <-- Pastikan import ini ada
)

// Interface service
type HemodialysisMonitoringService interface {
	CreateOrUpdateMonitoringForToday(userID uint, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error)
	GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error)
	GetMonitoringByID(userID, monitoringID uint) (models.HemodialysisMonitoring, error)
	// GetMonitoringByUserIDAndDate jika diperlukan
}

// Implementasi service
type hemodialysisMonitoringService struct {
	monitoringRepo repositories.HemodialysisMonitoringRepository
	userRepo       repositories.UserRepository // Opsional, hanya jika perlu timezone user di sini
}

// Constructor
func NewHemodialysisMonitoringService(monitoringRepo repositories.HemodialysisMonitoringRepository, userRepo repositories.UserRepository) HemodialysisMonitoringService {
	return &hemodialysisMonitoringService{
		monitoringRepo: monitoringRepo,
		userRepo:       userRepo,
	}
}

// CreateOrUpdateMonitoringForToday: Logika bisnis utama
func (s *hemodialysisMonitoringService) CreateOrUpdateMonitoringForToday(userID uint, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error) {
	// Gunakan UTC untuk konsistensi tanggal
	nowUTC := time.Now().UTC()
	todayUTC := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)

	// Cari apakah data untuk hari ini sudah ada
	existingMonitoring, err := s.monitoringRepo.FindByUserIDAndDate(userID, todayUTC)

	var savedMonitoring models.HemodialysisMonitoring
	var repoErr error

	if err != nil {
		// Jika error BUKAN karena tidak ditemukan, kembalikan error pencarian
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.HemodialysisMonitoring{}, fmt.Errorf("gagal mencari data monitoring hari ini: %w", err)
		}

		// --- BELUM ADA DATA HARI INI -> BUAT BARU ---
		newMonitoring := models.HemodialysisMonitoring{
			UserID:         userID,
			MonitoringDate: todayUTC,
			BPBefore:       input.BPBefore,
			BPAfter:        input.BPAfter,
			WeightBefore:   input.WeightBefore,
			WeightAfter:    input.WeightAfter,
		}
		savedMonitoring, repoErr = s.monitoringRepo.Create(newMonitoring)
		if repoErr != nil {
			return models.HemodialysisMonitoring{}, fmt.Errorf("gagal membuat data monitoring baru: %w", repoErr)
		}

	} else {
		// --- SUDAH ADA DATA HARI INI -> UPDATE ---
		existingMonitoring.BPBefore = input.BPBefore
		existingMonitoring.BPAfter = input.BPAfter
		existingMonitoring.WeightBefore = input.WeightBefore
		existingMonitoring.WeightAfter = input.WeightAfter

		savedMonitoring, repoErr = s.monitoringRepo.Update(existingMonitoring)
		if repoErr != nil {
			return models.HemodialysisMonitoring{}, fmt.Errorf("gagal memperbarui data monitoring: %w", repoErr)
		}
	}

	return savedMonitoring, nil
}

// GetMonitoringHistory mengambil riwayat
func (s *hemodialysisMonitoringService) GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error) {
	// Ambil 10 data terakhir
	history, err := s.monitoringRepo.FindHistoryByUserID(userID, 10)
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

	// Verifikasi kepemilikan
	if monitoring.UserID != userID {
		return models.HemodialysisMonitoring{}, errors.New("tidak berwenang mengakses data pemantauan ini")
	}

	return monitoring, nil
}
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
const DEMO_DATE_HEMODIALYSIS = "2026-01-14"

type HemodialysisMonitoringService interface {
	CreateOrUpdateMonitoringForToday(user models.User, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error)
	GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error)
	GetMonitoringByID(userID, monitoringID uint) (models.HemodialysisMonitoring, error)
}

type hemodialysisMonitoringService struct {
	monitoringRepo repositories.HemodialysisMonitoringRepository
	userRepo       repositories.UserRepository
}

func NewHemodialysisMonitoringService(monitoringRepo repositories.HemodialysisMonitoringRepository, userRepo repositories.UserRepository) HemodialysisMonitoringService {
	return &hemodialysisMonitoringService{
		monitoringRepo: monitoringRepo,
		userRepo:       userRepo,
	}
}

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

	// loc, err := time.LoadLocation(user.Timezone)
	// if err != nil {
	// 	return models.HemodialysisMonitoring{}, fmt.Errorf(
	// 		"invalid user timezone (%s): %w",
	// 		user.Timezone,
	// 		err,
	// 	)
	// }

	// ===============================
	// 2. ✅ DAPATKAN TANGGAL STRING (YYYY-MM-DD)
	// ===============================
	var todayStr string
	
	// PRODUCTION: Uncomment code di bawah
	// nowInUserTZ := time.Now().In(loc)
	// todayStr = nowInUserTZ.Format("2006-01-02")
	
	// ✅ DEMO MODE: Paksa tanggal
	todayStr = DEMO_DATE_HEMODIALYSIS
	
	// Parse ke time.Time UTC midnight untuk disimpan ke database
	todayDate, err := time.ParseInLocation("2006-01-02", todayStr, time.UTC)
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf("failed to parse date: %w", err)
	}

	// ===============================
	// 3. Validasi Input
	// ===============================
	if input.BPBefore == "" || input.BPAfter == "" {
		return models.HemodialysisMonitoring{}, fmt.Errorf("tekanan darah sebelum dan sesudah harus diisi")
	}
	if input.WeightBefore <= 0 || input.WeightAfter <= 0 {
		return models.HemodialysisMonitoring{}, fmt.Errorf("berat badan harus lebih dari 0")
	}

	// ===============================
	// 4. ✅ CARI DENGAN STRING DATE
	// ===============================
	existing, err := s.monitoringRepo.FindByUserIDAndDateString(user.ID, todayStr)

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
			MonitoringDate: todayDate, // ✅ UTC midnight
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

func (s *hemodialysisMonitoringService) GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error) {
	limit := 10
	if limit <= 0 || limit > 100 {
		limit = 10
	}

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

	if monitoring.UserID != userID {
		return models.HemodialysisMonitoring{}, errors.New("tidak berwenang mengakses data pemantauan ini")
	}

	return monitoring, nil
}
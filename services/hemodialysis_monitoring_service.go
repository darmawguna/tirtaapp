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
	CreateOrUpdateMonitoringForToday(user models.User, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error)
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
func (s *hemodialysisMonitoringService) CreateOrUpdateMonitoringForToday(
	user models.User,
	input dto.CreateHemodialysisMonitoringDTO,
) (models.HemodialysisMonitoring, error) {

	// 1. Load timezone user
	loc, err := time.LoadLocation(user.Timezone)
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf(
			"invalid user timezone (%s): %w",
			user.Timezone,
			err,
		)
	}

	// 2. Hitung "hari ini" berdasarkan timezone user
	nowUser := time.Now().In(loc)
	todayUser := time.Date(
		nowUser.Year(),
		nowUser.Month(),
		nowUser.Day(),
		0, 0, 0, 0,
		loc,
	)

	existing, err := s.monitoringRepo.FindByUserIDAndDate(
		user.ID,
		todayUser,
	)

	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.HemodialysisMonitoring{}, fmt.Errorf(
				"failed to check today's monitoring: %w",
				err,
			)
		}

		// --- BELUM ADA → CREATE ---
		newMonitoring := models.HemodialysisMonitoring{
			UserID:         user.ID,
			MonitoringDate: todayUser,
			BPBefore:       input.BPBefore,
			BPAfter:        input.BPAfter,
			WeightBefore:   input.WeightBefore,
			WeightAfter:    input.WeightAfter,
		}

		return s.monitoringRepo.Create(newMonitoring)
	}

	// --- SUDAH ADA → UPDATE ---
	existing.BPBefore = input.BPBefore
	existing.BPAfter = input.BPAfter
	existing.WeightBefore = input.WeightBefore
	existing.WeightAfter = input.WeightAfter

	return s.monitoringRepo.Update(existing)
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
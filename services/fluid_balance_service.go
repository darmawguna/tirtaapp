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

// Definisikan batas cairan (bisa diambil dari config nantinya)
const dailyIntakeLimit = 600
const warningThreshold = 500

// ✅ DEMO MODE: Hard-code tanggal untuk sidang
const DEMO_DATE = "2026-01-14" // Ubah ini setelah sidang selesai

type FluidBalanceService interface {
	CreateOrUpdateLog(user models.User, input dto.CreateOrUpdateFluidLogDTO) (models.FluidBalanceLog, error)
	GetUserHistory(userID uint) ([]models.FluidBalanceLog, error)
}

type fluidBalanceService struct {
	repo     repositories.FluidBalanceRepository
	userRepo repositories.UserRepository
}

func NewFluidBalanceService(repo repositories.FluidBalanceRepository, userRepo repositories.UserRepository) FluidBalanceService {
	return &fluidBalanceService{repo: repo, userRepo: userRepo}
}

func (s *fluidBalanceService) CreateOrUpdateLog(
	user models.User,
	input dto.CreateOrUpdateFluidLogDTO,
) (models.FluidBalanceLog, error) {

	// ===============================
	// 1. Validasi dan Load Timezone User
	// ===============================
	if user.Timezone == "" {
		return models.FluidBalanceLog{}, fmt.Errorf("user timezone not set")
	}
	
	// loc, err := time.LoadLocation(user.Timezone)
	// if err != nil {
	// 	return models.FluidBalanceLog{}, fmt.Errorf("invalid user timezone: %s", user.Timezone)
	// }

	// ===============================
	// 2. ✅ GUNAKAN TANGGAL DEMO UNTUK SIDANG
	// ===============================
	// PRODUCTION: Uncomment code di bawah dan hapus DEMO_DATE
	// nowInUserTZ := time.Now().In(loc)
	// todayStr := nowInUserTZ.Format("2006-01-02")
	
	// ✅ DEMO MODE: Paksa tanggal ke 14 Januari 2026
	todayStr := DEMO_DATE
	
	// Parse kembali ke time.Time untuk disimpan ke database (UTC midnight)
	todayDate, err := time.Parse("2006-01-02", todayStr)
	if err != nil {
		return models.FluidBalanceLog{}, fmt.Errorf("failed to parse date: %w", err)
	}

	// ===============================
	// 3. Ambil Nilai Intake/Output
	// ===============================
	intakeVal := 0
	outputVal := 0

	if input.IntakeCC != nil {
		intakeVal = *input.IntakeCC
	}
	if input.OutputCC != nil {
		outputVal = *input.OutputCC
	}

	// ===============================
	// 4. Cari Log Hari Ini (berdasarkan string date)
	// ===============================
	existingLog, err := s.repo.FindByUserAndDate(user.ID, todayStr)

	// ===============================
	// 5A. CREATE - Jika Log Belum Ada
	// ===============================
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.FluidBalanceLog{}, fmt.Errorf("failed to query log: %w", err)
		}

		newLog := models.FluidBalanceLog{
			UserID:   user.ID,
			LogDate:  todayDate, // ✅ Akan selalu 2026-01-14
			IntakeCC: intakeVal,
			OutputCC: outputVal,
		}

		newLog.BalanceCC = newLog.IntakeCC - newLog.OutputCC

		if newLog.BalanceCC >= warningThreshold {
			newLog.WarningMessage = fmt.Sprintf(
				"Peringatan!\n\nTotal keseimbangan cairan hari ini (%d cc) mendekati batas (%d cc/24 jam).",
				newLog.BalanceCC,
				dailyIntakeLimit,
			)
		}

		return s.repo.Create(newLog)
	}

	// ===============================
	// 5B. UPDATE - Jika Log Sudah Ada
	// ===============================
	existingLog.IntakeCC += intakeVal
	existingLog.OutputCC += outputVal
	existingLog.BalanceCC = existingLog.IntakeCC - existingLog.OutputCC
	existingLog.WarningMessage = ""

	if existingLog.BalanceCC >= warningThreshold {
		existingLog.WarningMessage = fmt.Sprintf(
			"Peringatan!\n\nTotal keseimbangan cairan hari ini (%d cc) mendekati batas (%d cc/24 jam).",
			existingLog.BalanceCC,
			dailyIntakeLimit,
		)
	}

	return s.repo.Update(existingLog)
}

func (s *fluidBalanceService) GetUserHistory(userID uint) ([]models.FluidBalanceLog, error) {
	logs, err := s.repo.FindHistoryByUserID(userID, 7)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat cairan: %w", err)
	}
	return logs, nil
}
package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto" // Sesuaikan path
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm"
)

// Definisikan batas cairan (bisa diambil dari config nantinya)
const dailyIntakeLimit = 600
const warningThreshold = 500

type FluidBalanceService interface {
	CreateOrUpdateLog(userID uint, input dto.CreateOrUpdateFluidLogDTO) (models.FluidBalanceLog, error)
	GetUserHistory(userID uint) ([]models.FluidBalanceLog, error)
}

type fluidBalanceService struct {
	repo     repositories.FluidBalanceRepository
	userRepo repositories.UserRepository
}

func NewFluidBalanceService(repo repositories.FluidBalanceRepository, userRepo repositories.UserRepository) FluidBalanceService {
	return &fluidBalanceService{repo: repo, userRepo: userRepo}
}

func (s *fluidBalanceService) CreateOrUpdateLog(userID uint, input dto.CreateOrUpdateFluidLogDTO) (models.FluidBalanceLog, error) {
	nowUTC := time.Now().UTC()
	todayUTC := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)

	// [LOGIKA BISNIS] Cek apakah log sudah ada untuk hari ini
	existingLog, err := s.repo.FindByUserAndDate(userID, todayUTC)

	var finalLog models.FluidBalanceLog
	var repoErr error

	var intakeVal, outputVal int
    if input.IntakeCC != nil { // Cek nil sebelum dereference
        intakeVal = *input.IntakeCC
    }
    if input.OutputCC != nil { // Cek nil sebelum dereference
        outputVal = *input.OutputCC
    }
	
	if err != nil {
		// Jika error BUKAN karena tidak ditemukan, kembalikan error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.FluidBalanceLog{}, fmt.Errorf("gagal mencari log hari ini: %w", err)
		}

		// --- Kasus 1: Record belum ada -> Buat Baru ---
		log.Println("No existing log found for today. Creating new one.")
		newLog := models.FluidBalanceLog{
			UserID:   userID,
			LogDate:  todayUTC,
			IntakeCC: intakeVal,
			OutputCC: outputVal,
		}
		newLog.BalanceCC = newLog.IntakeCC - newLog.OutputCC
		// Terapkan warning jika perlu
		if newLog.BalanceCC >= warningThreshold {
			newLog.WarningMessage = fmt.Sprintf("Peringatan!\n\nHalo Bapak/Ibu, total keseimbangan cairan Anda hari ini (%d cc) sudah mendekati batas maksimal harian (%d cc/24 jam). Ingat, kelebihan cairan bisa menimbulkan sesak napas dan bengkak. Mari jaga kesehatan dengan mematuhi batas cairan harian Anda. Informasi lengkap tentang pengelolaan cairan dapat dilihat di menu Edukasi.", newLog.BalanceCC, dailyIntakeLimit)
			log.Printf("Warning triggered for user %d, accumulated balance: %d", userID, newLog.BalanceCC)
		}
		// Panggil repo.Create
		finalLog, repoErr = s.repo.Create(newLog)

	} else {
		// --- Kasus 2: Record sudah ada -> Update ---
		log.Println("Existing log found for today (ID:", existingLog.ID, "). Updating.")
		// Akumulasi nilai
		existingLog.IntakeCC += intakeVal
		existingLog.OutputCC += outputVal
		// Hitung ulang balance & warning
		existingLog.BalanceCC = existingLog.IntakeCC - existingLog.OutputCC
		existingLog.WarningMessage = "" // Reset warning
		if existingLog.BalanceCC >= warningThreshold {
			existingLog.WarningMessage = fmt.Sprintf("Peringatan!\n\nHalo Bapak/Ibu, total keseimbangan cairan Anda hari ini (%d cc) sudah mendekati batas maksimal harian (%d cc/24 jam). Ingat, kelebihan cairan bisa menimbulkan sesak napas dan bengkak. Mari jaga kesehatan dengan mematuhi batas cairan harian Anda. Informasi lengkap tentang pengelolaan cairan dapat dilihat di menu Edukasi.", existingLog.BalanceCC, dailyIntakeLimit)
			log.Printf("Warning triggered...")
		}
		// Panggil repo.Update
		finalLog, repoErr = s.repo.Update(existingLog)
	}

	// Tangani error dari operasi Create atau Update
	if repoErr != nil {
		return models.FluidBalanceLog{}, fmt.Errorf("gagal menyimpan log cairan: %w", repoErr)
	}

	return finalLog, nil
}

func (s *fluidBalanceService) GetUserHistory(userID uint) ([]models.FluidBalanceLog, error) {
	logs, err := s.repo.FindHistoryByUserID(userID, 7)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil riwayat cairan: %w", err)
	}
	return logs, nil
}

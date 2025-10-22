package services

import (
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
const warningThreshold = 500 // Batas balance untuk memicu warning

type FluidBalanceService interface {
	CreateOrUpdateLog(userID uint, input dto.CreateOrUpdateFluidLogDTO) (models.FluidBalanceLog, error)
	GetUserHistory(userID uint) ([]models.FluidBalanceLog, error)
}

type fluidBalanceService struct {
	userRepo repositories.UserRepository
	repo repositories.FluidBalanceRepository
}

func NewFluidBalanceService(repo repositories.FluidBalanceRepository, userRepo repositories.UserRepository) FluidBalanceService {
	return &fluidBalanceService{repo: repo, userRepo: userRepo}
}


func (s *fluidBalanceService) CreateOrUpdateLog(userID uint, input dto.CreateOrUpdateFluidLogDTO) (models.FluidBalanceLog, error) {
	// [PEMBARUAN] Ambil data user untuk mendapatkan timezone
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return models.FluidBalanceLog{}, fmt.Errorf("could not find user %d: %w", userID, err)
	}

	// Muat lokasi/zona waktu spesifik milik user
	location, err := time.LoadLocation(user.Timezone)
	if err != nil {
		log.Printf("WARNING: Invalid timezone '%s' for user %d. Falling back to UTC.", user.Timezone, user.ID)
		location, _ = time.LoadLocation("UTC") // Fallback ke UTC jika timezone tidak valid
	}

	now := time.Now().In(location) // Dapatkan waktu saat ini di zona waktu user
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location) // Tanggal hari ini di zona waktu user

	// Cek apakah sudah ada log untuk hari ini
	existingLog, err := s.repo.FindByUserAndDate(userID, today)
	if err != nil && err != gorm.ErrRecordNotFound {
		return models.FluidBalanceLog{}, err
	}

	// Buat atau update log
	logEntry := existingLog
	logEntry.UserID = userID
	logEntry.LogDate = today // Simpan tanggal sesuai zona waktu user

	// Logika Akumulasi (tidak berubah)
	if logEntry.ID == 0 {
		logEntry.IntakeCC = input.IntakeCC
		logEntry.OutputCC = input.OutputCC
	} else {
		logEntry.IntakeCC += input.IntakeCC
		logEntry.OutputCC += input.OutputCC
	}

	// Hitung balance berdasarkan total akumulasi
	logEntry.BalanceCC = logEntry.IntakeCC - logEntry.OutputCC
	logEntry.WarningMessage = "" // Reset warning message

	// Terapkan aturan warning
	if logEntry.BalanceCC >= warningThreshold {
		logEntry.WarningMessage = fmt.Sprintf("Peringatan!\n\nHalo Bapak/Ibu, total keseimbangan cairan Anda hari ini (%d cc) sudah mendekati batas maksimal harian (%d cc/24 jam). Ingat, kelebihan cairan bisa menimbulkan sesak napas dan bengkak. Mari jaga kesehatan dengan mematuhi batas cairan harian Anda. Informasi lengkap tentang pengelolaan cairan dapat dilihat di menu Edukasi.", logEntry.BalanceCC, dailyIntakeLimit)
		log.Printf("Warning triggered for user %d, accumulated balance: %d", userID, logEntry.BalanceCC)
	}

	// Simpan ke database (Upsert)
	return s.repo.Upsert(logEntry)
}

func (s *fluidBalanceService) GetUserHistory(userID uint) ([]models.FluidBalanceLog, error) {
	// Ambil riwayat, misalnya 7 hari terakhir
	return s.repo.FindHistoryByUserID(userID, 7)
}
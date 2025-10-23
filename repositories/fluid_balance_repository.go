package repositories

import (
	// "errors" // Tidak perlu lagi di sini
	// "fmt" // Tidak perlu lagi di sini
	"time"

	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type FluidBalanceRepository interface {
	FindByUserAndDate(userID uint, date time.Time) (models.FluidBalanceLog, error)
	// Hapus Upsert
	// Upsert(log models.FluidBalanceLog) (models.FluidBalanceLog, error)
	Create(log models.FluidBalanceLog) (models.FluidBalanceLog, error) // <-- Tambah Create
	Update(log models.FluidBalanceLog) (models.FluidBalanceLog, error) // <-- Tambah Update
	FindHistoryByUserID(userID uint, limit int) ([]models.FluidBalanceLog, error)
}

type fluidBalanceRepository struct {
	db *gorm.DB
}

func NewFluidBalanceRepository(db *gorm.DB) FluidBalanceRepository {
	return &fluidBalanceRepository{db: db}
}

// FindByUserAndDate: Tetap sama, gunakan DATE() SQL
func (r *fluidBalanceRepository) FindByUserAndDate(userID uint, date time.Time) (models.FluidBalanceLog, error) {
	var log models.FluidBalanceLog
	err := r.db.Where("user_id = ? AND DATE(log_date) = DATE(?)", userID, date.UTC()).First(&log).Error
	return log, err
}

// [BARU] Create: Fungsi sederhana untuk INSERT
func (r *fluidBalanceRepository) Create(log models.FluidBalanceLog) (models.FluidBalanceLog, error) {
	err := r.db.Create(&log).Error
	return log, err
}

// [BARU] Update: Fungsi sederhana untuk UPDATE
func (r *fluidBalanceRepository) Update(log models.FluidBalanceLog) (models.FluidBalanceLog, error) {
	// Gunakan Save karena log sudah memiliki ID yang valid
	err := r.db.Save(&log).Error
	return log, err
}


func (r *fluidBalanceRepository) FindHistoryByUserID(userID uint, limit int) ([]models.FluidBalanceLog, error) {
	var logs []models.FluidBalanceLog
	err := r.db.Where("user_id = ?", userID).Order("log_date desc").Limit(limit).Find(&logs).Error
	return logs, err
}
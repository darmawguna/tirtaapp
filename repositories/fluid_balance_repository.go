package repositories

import (
	"time"

	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path
	"gorm.io/gorm"
)

type FluidBalanceRepository interface {
	// Mencari log berdasarkan userID dan tanggal.
	FindByUserAndDate(userID uint, date time.Time) (models.FluidBalanceLog, error)
	// Membuat atau memperbarui log (jika sudah ada untuk tanggal tersebut).
	Upsert(log models.FluidBalanceLog) (models.FluidBalanceLog, error)
	// Mengambil riwayat log untuk user.
	FindHistoryByUserID(userID uint, limit int) ([]models.FluidBalanceLog, error)
}

type fluidBalanceRepository struct {
	db *gorm.DB
}

func NewFluidBalanceRepository(db *gorm.DB) FluidBalanceRepository {
	return &fluidBalanceRepository{db: db}
}

func (r *fluidBalanceRepository) FindByUserAndDate(userID uint, date time.Time) (models.FluidBalanceLog, error) {
	var log models.FluidBalanceLog
	// Hanya ambil tanggalnya saja
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	err := r.db.Where("user_id = ? AND log_date = ?", userID, dateOnly).First(&log).Error
	return log, err
}

func (r *fluidBalanceRepository) Upsert(log models.FluidBalanceLog) (models.FluidBalanceLog, error) {
	// GORM's Save akan INSERT jika ID 0, atau UPDATE jika ID sudah ada.
	err := r.db.Save(&log).Error
	return log, err
}

func (r *fluidBalanceRepository) FindHistoryByUserID(userID uint, limit int) ([]models.FluidBalanceLog, error) {
	var logs []models.FluidBalanceLog
	err := r.db.Where("user_id = ?", userID).Order("log_date desc").Limit(limit).Find(&logs).Error
	return logs, err
}
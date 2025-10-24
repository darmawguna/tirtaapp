package repositories

import (
	"fmt"
	"time"

	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path jika berbeda
	"gorm.io/gorm"
)

// Interface untuk HemodialysisMonitoringRepository
type HemodialysisMonitoringRepository interface {
	FindByUserIDAndDate(userID uint, date time.Time) (models.HemodialysisMonitoring, error)
	Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error)
	Update(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error)
	FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error)
	FindByID(id uint) (models.HemodialysisMonitoring, error)
}

// Implementasi repository
type hemodialysisMonitoringRepository struct {
	db *gorm.DB
}

// Constructor
func NewHemodialysisMonitoringRepository(db *gorm.DB) HemodialysisMonitoringRepository {
	return &hemodialysisMonitoringRepository{db: db}
}

// FindByUserIDAndDate mencari data monitoring berdasarkan user dan tanggal (UTC)
func (r *hemodialysisMonitoringRepository) FindByUserIDAndDate(userID uint, date time.Time) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	// Gunakan DATE() SQL untuk perbandingan tanggal yang andal
	err := r.db.Where("user_id = ? AND DATE(monitoring_date) = DATE(?)", userID, date.UTC()).First(&monitoring).Error
	return monitoring, err
}

// Create: Fungsi sederhana untuk INSERT
func (r *hemodialysisMonitoringRepository) Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error) {
	err := r.db.Create(&monitoring).Error
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf("gagal create monitoring: %w", err)
	}
	return monitoring, nil
}

// Update: Fungsi sederhana untuk UPDATE
func (r *hemodialysisMonitoringRepository) Update(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error) {
	// Gunakan Save karena monitoring sudah memiliki ID yang valid
	err := r.db.Save(&monitoring).Error
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf("gagal update monitoring: %w", err)
	}
	return monitoring, nil
}

// FindHistoryByUserID mengambil riwayat
func (r *hemodialysisMonitoringRepository) FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error) {
	var monitorings []models.HemodialysisMonitoring
	err := r.db.Where("user_id = ?", userID).Order("monitoring_date desc").Limit(limit).Find(&monitorings).Error
	return monitorings, err
}

func (r *hemodialysisMonitoringRepository) FindByID(id uint) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	err := r.db.First(&monitoring, id).Error // Cari berdasarkan Primary Key 'id'
	return monitoring, err
}
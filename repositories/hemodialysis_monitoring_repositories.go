package repositories

import (
	"time"

	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path jika berbeda
	"gorm.io/gorm"
)

// Interface untuk HemodialysisMonitoringRepository
type HemodialysisMonitoringRepository interface {
	FindByUserIDAndDate(userID uint, date time.Time) (models.HemodialysisMonitoring, error)
	FindByUserIDAndDateString(userID uint, dateStr string) (models.HemodialysisMonitoring, error) // ✅ TAMBAHKAN INI
	Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error)
	Update(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error)
	FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error)
	FindByID(id uint) (models.HemodialysisMonitoring, error)
}

type hemodialysisMonitoringRepository struct {
	db *gorm.DB
}

func NewHemodialysisMonitoringRepository(db *gorm.DB) HemodialysisMonitoringRepository {
	return &hemodialysisMonitoringRepository{db: db}
}

// Method lama (tetap ada untuk backward compatibility)
func (r *hemodialysisMonitoringRepository) FindByUserIDAndDate(
	userID uint,
	date time.Time,
) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	err := r.db.
		Where("user_id = ? AND DATE(monitoring_date) = DATE(?)", userID, date).
		First(&monitoring).Error
	return monitoring, err
}

// ✅ METHOD BARU: Query dengan string date
func (r *hemodialysisMonitoringRepository) FindByUserIDAndDateString(
	userID uint,
	dateStr string, // Format: "2006-01-02"
) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	
	// ✅ Gunakan DATE() function untuk mencocokkan hanya tanggal
	err := r.db.
		Where("user_id = ? AND DATE(monitoring_date) = ?", userID, dateStr).
		First(&monitoring).Error
	
	return monitoring, err
}

func (r *hemodialysisMonitoringRepository) Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error) {
	err := r.db.Create(&monitoring).Error
	return monitoring, err
}

func (r *hemodialysisMonitoringRepository) Update(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error) {
	err := r.db.Save(&monitoring).Error
	return monitoring, err
}

func (r *hemodialysisMonitoringRepository) FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error) {
	var monitorings []models.HemodialysisMonitoring
	
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	
	err := r.db.
		Where("user_id = ?", userID).
		Order("monitoring_date DESC").
		Limit(limit).
		Find(&monitorings).Error
	
	return monitorings, err
}

func (r *hemodialysisMonitoringRepository) FindByID(id uint) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	err := r.db.First(&monitoring, id).Error
	return monitoring, err
}
package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path
	"gorm.io/gorm"
)

type HemodialysisMonitoringRepository interface {
	Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error)
	FindByScheduleID(scheduleID uint) (models.HemodialysisMonitoring, error)
	FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error)
}

type hemodialysisMonitoringRepository struct {
	db *gorm.DB
}

func NewHemodialysisMonitoringRepository(db *gorm.DB) HemodialysisMonitoringRepository {
	return &hemodialysisMonitoringRepository{db: db}
}

func (r *hemodialysisMonitoringRepository) Create(monitoring models.HemodialysisMonitoring) (models.HemodialysisMonitoring, error) {
	// Using FirstOrCreate to prevent duplicate entries for the same schedule ID
	err := r.db.Where(models.HemodialysisMonitoring{HemodialysisScheduleID: monitoring.HemodialysisScheduleID}).
		Attrs(monitoring).               // Attributes only used if record is created
		FirstOrCreate(&monitoring).Error // Finds or creates based on ScheduleID
	return monitoring, err
}

func (r *hemodialysisMonitoringRepository) FindByScheduleID(scheduleID uint) (models.HemodialysisMonitoring, error) {
	var monitoring models.HemodialysisMonitoring
	// Preload the schedule to easily get the date later
	err := r.db.Preload("HemodialysisSchedule").Where("hemodialysis_schedule_id = ?", scheduleID).First(&monitoring).Error
	return monitoring, err
}

func (r *hemodialysisMonitoringRepository) FindHistoryByUserID(userID uint, limit int) ([]models.HemodialysisMonitoring, error) {
	var monitorings []models.HemodialysisMonitoring
	// Preload schedule and order by its date
	err := r.db.Preload("HemodialysisSchedule").
		Joins("JOIN hemodialysis_schedules on hemodialysis_schedules.id = hemodialysis_monitorings.hemodialysis_schedule_id").
		Where("hemodialysis_monitorings.user_id = ?", userID).
		Order("hemodialysis_schedules.schedule_date desc").
		Limit(limit).
		Find(&monitorings).Error
	return monitorings, err
}

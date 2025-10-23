package repositories

import (
	"time"

	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type HemodialysisScheduleRepository interface {
	Create(schedule models.HemodialysisSchedule) (models.HemodialysisSchedule, error)
	FindAllByUserID(userID uint) ([]models.HemodialysisSchedule, error)
	FindByID(id uint) (models.HemodialysisSchedule, error)
	Update(schedule models.HemodialysisSchedule) (models.HemodialysisSchedule, error)
	FindSchedulesForDateAndNotNotified(date time.Time) ([]models.HemodialysisSchedule, error)
	Delete(id uint) error
}

type hemodialysisScheduleRepository struct {
	db *gorm.DB
}

func NewHemodialysisScheduleRepository(db *gorm.DB) HemodialysisScheduleRepository {
	return &hemodialysisScheduleRepository{db: db}
}

func (r *hemodialysisScheduleRepository) Create(schedule models.HemodialysisSchedule) (models.HemodialysisSchedule, error) {
	err := r.db.Create(&schedule).Error
	return schedule, err
}

func (r *hemodialysisScheduleRepository) FindAllByUserID(userID uint) ([]models.HemodialysisSchedule, error) {
	var schedules []models.HemodialysisSchedule
	err := r.db.Where("user_id = ?", userID).Order("schedule_date desc").Find(&schedules).Error
	return schedules, err
}

func (r *hemodialysisScheduleRepository) FindByID(id uint) (models.HemodialysisSchedule, error) {
	var schedule models.HemodialysisSchedule
	err := r.db.First(&schedule, id).Error
	return schedule, err
}

func (r *hemodialysisScheduleRepository) FindSchedulesForDateAndNotNotified(date time.Time) ([]models.HemodialysisSchedule, error) {
	var schedules []models.HemodialysisSchedule
	err := r.db.Where("schedule_date = ? AND is_active = ? AND monitoring_notification_sent = ?", date, true, false).Find(&schedules).Error
	return schedules, err
}

func (r *hemodialysisScheduleRepository) Update(schedule models.HemodialysisSchedule) (models.HemodialysisSchedule, error) {
	err := r.db.Save(&schedule).Error
	return schedule, err
}

func (r *hemodialysisScheduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.HemodialysisSchedule{}, id).Error
}
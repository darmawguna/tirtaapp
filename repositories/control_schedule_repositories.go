package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type ControlScheduleRepository interface {
	Create(schedule models.ControlSchedule) (models.ControlSchedule, error)
	FindAllByUserID(userID uint) ([]models.ControlSchedule, error)
	FindByID(id uint) (models.ControlSchedule, error)
	Update(schedule models.ControlSchedule) (models.ControlSchedule, error)
	Delete(id uint) error
}

type controlScheduleRepository struct {
	db *gorm.DB
}

func NewControlScheduleRepository(db *gorm.DB) ControlScheduleRepository {
	return &controlScheduleRepository{db: db}
}

func (r *controlScheduleRepository) Create(schedule models.ControlSchedule) (models.ControlSchedule, error) {
	err := r.db.Create(&schedule).Error
	return schedule, err
}

func (r *controlScheduleRepository) FindAllByUserID(userID uint) ([]models.ControlSchedule, error) {
	var schedules []models.ControlSchedule
	err := r.db.Where("user_id = ?", userID).Order("control_date desc").Find(&schedules).Error
	return schedules, err
}

func (r *controlScheduleRepository) FindByID(id uint) (models.ControlSchedule, error) {
	var schedule models.ControlSchedule
	err := r.db.First(&schedule, id).Error
	return schedule, err
}

func (r *controlScheduleRepository) Update(schedule models.ControlSchedule) (models.ControlSchedule, error) {
	err := r.db.Save(&schedule).Error
	return schedule, err
}

func (r *controlScheduleRepository) Delete(id uint) error {
	return r.db.Delete(&models.ControlSchedule{}, id).Error
}
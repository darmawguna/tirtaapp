package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type DrugScheduleRepository interface {
	Create(schedule models.DrugSchedule) (models.DrugSchedule, error)
	FindAllByUserID(userID uint) ([]models.DrugSchedule, error)
	FindByID(id uint) (models.DrugSchedule, error)
	Update(schedule models.DrugSchedule) (models.DrugSchedule, error)
	Delete(id uint) error
}

type drugScheduleRepository struct {
	db *gorm.DB
}

func NewDrugScheduleRepository(db *gorm.DB) DrugScheduleRepository {
	return &drugScheduleRepository{db: db}
}

func (r *drugScheduleRepository) Create(schedule models.DrugSchedule) (models.DrugSchedule, error) {
	err := r.db.Create(&schedule).Error
	return schedule, err
}

func (r *drugScheduleRepository) FindAllByUserID(userID uint) ([]models.DrugSchedule, error) {
	var schedules []models.DrugSchedule
	// Mengurutkan berdasarkan tanggal terbaru
	err := r.db.Where("user_id = ?", userID).Order("schedule_date desc").Find(&schedules).Error
	return schedules, err
}

func (r *drugScheduleRepository) FindByID(id uint) (models.DrugSchedule, error) {
	var schedule models.DrugSchedule
	err := r.db.First(&schedule, id).Error
	return schedule, err
}

func (r *drugScheduleRepository) Update(schedule models.DrugSchedule) (models.DrugSchedule, error) {
	err := r.db.Save(&schedule).Error
	return schedule, err
}

func (r *drugScheduleRepository) Delete(id uint) error {
	var schedule models.DrugSchedule
	// Pastikan user hanya bisa menghapus jadwal miliknya (opsional, bisa juga di service)
	// Untuk saat ini, kita sederhanakan.
	return r.db.Delete(&schedule, id).Error
}
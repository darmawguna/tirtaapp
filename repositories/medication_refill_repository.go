package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path
	"gorm.io/gorm"
)

type MedicationRefillRepository interface {
	Create(schedule models.MedicationRefillSchedule) (models.MedicationRefillSchedule, error)
	FindAllByUserID(userID uint) ([]models.MedicationRefillSchedule, error)
	FindByID(id uint) (models.MedicationRefillSchedule, error)
	Update(schedule models.MedicationRefillSchedule) (models.MedicationRefillSchedule, error)
	Delete(id uint) error
}

type medicationRefillRepository struct {
	db *gorm.DB
}

func NewMedicationRefillRepository(db *gorm.DB) MedicationRefillRepository {
	return &medicationRefillRepository{db: db}
}

func (r *medicationRefillRepository) Create(medicationRefillSchedule models.MedicationRefillSchedule) (models.MedicationRefillSchedule, error) {
	err := r.db.Create(&medicationRefillSchedule).Error
	return medicationRefillSchedule, err
}

func (r *medicationRefillRepository) FindAllByUserID(userID uint) ([]models.MedicationRefillSchedule, error) {
	var medicationRefillSchedule []models.MedicationRefillSchedule
	err := r.db.Where("user_id = ?", userID).Order("refill_date desc").Find(&medicationRefillSchedule).Error
	return medicationRefillSchedule, err
}

func (r *medicationRefillRepository) FindByID(id uint) (models.MedicationRefillSchedule, error) {
	var medicationRefillSchedule models.MedicationRefillSchedule
	err := r.db.First(&medicationRefillSchedule, id).Error
	return medicationRefillSchedule, err
}

func (r *medicationRefillRepository) Update(medicationRefillSchedule models.MedicationRefillSchedule) (models.MedicationRefillSchedule, error) {
	err := r.db.Save(&medicationRefillSchedule).Error
	return medicationRefillSchedule, err
}

func (r *medicationRefillRepository) Delete(id uint) error {
	return r.db.Delete(&models.MedicationRefillSchedule{}, id).Error
}
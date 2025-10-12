package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type EducationRepository interface {
	Create(education models.Education) (models.Education, error)
	FindAll() ([]models.Education, error)
	FindByID(id uint) (models.Education, error)
	Update(education models.Education) (models.Education, error)
	Delete(id uint) error
}

type educationRepository struct {
	db *gorm.DB
}

func NewEducationRepository(db *gorm.DB) EducationRepository {
	return &educationRepository{db: db}
}

func (r *educationRepository) Create(education models.Education) (models.Education, error) {
	err := r.db.Create(&education).Error
	return education, err
}

func (r *educationRepository) FindAll() ([]models.Education, error) {
	var educations []models.Education
	err := r.db.Find(&educations).Error
	return educations, err
}

func (r *educationRepository) FindByID(id uint) (models.Education, error) {
	var education models.Education
	err := r.db.First(&education, id).Error
	return education, err
}

func (r *educationRepository) Update(education models.Education) (models.Education, error) {
	err := r.db.Save(&education).Error
	return education, err
}

func (r *educationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Education{}, id).Error
}
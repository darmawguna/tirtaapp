package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type ComplaintRepository interface {
	Create(log models.HemodialysisComplaint) (models.HemodialysisComplaint, error)

	FindByUserID(userID uint) ([]models.HemodialysisComplaint, error)
	FindByUserIDAndPhase(userID uint, phase string) ([]models.HemodialysisComplaint, error)

	FindByIDAndUserID(id uint, userID uint) (models.HemodialysisComplaint, error)
}

type complaintRepository struct {
	db *gorm.DB
}

func NewComplaintRepository(db *gorm.DB) ComplaintRepository {
	return &complaintRepository{db: db}
}

func (r *complaintRepository) Create(log models.HemodialysisComplaint) (models.HemodialysisComplaint, error) {
	err := r.db.Create(&log).Error
	return log, err
}

func (r *complaintRepository) FindByUserID(userID uint) ([]models.HemodialysisComplaint, error) {
	var logs []models.HemodialysisComplaint
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&logs).Error
	return logs, err
}

func (r *complaintRepository) FindByUserIDAndPhase(userID uint, phase string) ([]models.HemodialysisComplaint, error) {
	var logs []models.HemodialysisComplaint
	err := r.db.Where("user_id = ? AND phase = ?", userID, phase).Order("created_at desc").Find(&logs).Error
	return logs, err
}

func (r *complaintRepository) FindByIDAndUserID(id uint, userID uint) (models.HemodialysisComplaint, error) {
	var log models.HemodialysisComplaint
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&log).Error
	return log, err
}

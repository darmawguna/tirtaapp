package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type ComplaintRepository interface {
	Create(log models.ComplaintLog) (models.ComplaintLog, error)
	FindByUserID(userID uint) ([]models.ComplaintLog, error)
	FindByID(complaint_id uint) (models.ComplaintLog, error)
}

type complaintRepository struct {
	db *gorm.DB
}

func NewComplaintRepository(db *gorm.DB) ComplaintRepository {
	return &complaintRepository{db: db}
}

func (r *complaintRepository) Create(log models.ComplaintLog) (models.ComplaintLog, error) {
	err := r.db.Create(&log).Error
	return log, err
}

func (r *complaintRepository) FindByUserID(userID uint) ([]models.ComplaintLog, error) {
	var logs []models.ComplaintLog
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&logs).Error
	return logs, err
}

func (r *complaintRepository) FindByID(complaint_id uint) (models.ComplaintLog, error) {
	var logs models.ComplaintLog
	err := r.db.First(&logs, complaint_id).Error
	return logs, err
}
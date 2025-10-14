package services

import (
	"encoding/json"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/datatypes"
)

type ComplaintService interface {
	ProcessComplaint(userID uint, input dto.CreateComplaintDTO) (string, error)
	GetMyComplaints(userID uint) ([]models.ComplaintLog, error)
	GetComplainById (complaint_id uint) (models.ComplaintLog, error)
}

type complaintService struct {
	complaintRepo repositories.ComplaintRepository
}

func NewComplaintService(complaintRepo repositories.ComplaintRepository) ComplaintService {
	return &complaintService{complaintRepo: complaintRepo}
}

func (s *complaintService) ProcessComplaint(userID uint, input dto.CreateComplaintDTO) (string, error) {
	var generatedMessage string
	complaintCount := len(input.Complaints)

	// Logika bisnis utama
	if complaintCount == 1 {
		generatedMessage = "Konsultasikan keluhan bapak/ibu kepada dokter/perawat yang bertugas atau hubungi petugas pada link TANYA PETUGAS"
	} else if complaintCount > 1 {
		generatedMessage = "Segera konsultasikan keluhan bapak/ibu ke poliklinik atau faskes terdekat"
	}

	// Ubah array string menjadi format JSON untuk disimpan
	complaintsJSON, err := json.Marshal(input.Complaints)
	if err != nil {
		return "", err
	}

	// Buat log untuk disimpan
	log := models.ComplaintLog{
		UserID:     userID,
		Complaints: datatypes.JSON(complaintsJSON),
		Message:    generatedMessage,
	}

	// Simpan log ke database
	_, err = s.complaintRepo.Create(log)
	if err != nil {
		return "", err
	}

	return generatedMessage, nil
}

func (s *complaintService) GetMyComplaints(userID uint) ([]models.ComplaintLog, error) {
	return s.complaintRepo.FindByUserID(userID)
}

func (s *complaintService) GetComplainById(complaint_id uint) (models.ComplaintLog, error) {
	return s.complaintRepo.FindByID(complaint_id)
}
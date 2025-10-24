package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm"
)

type EducationService interface {
	Create(input dto.CreateEducationDTO, createdBy uint, thumbnailPath string) (models.Education, error)
	FindAll() ([]models.Education, error)
	FindByID(id uint) (models.Education, error)
	Update(id uint, input dto.UpdateEducationDTO, thumbnailPath *string) (models.Education, error)
	Delete(id uint) error
}

type educationService struct {
	educationRepo repositories.EducationRepository
}

func NewEducationService(educationRepo repositories.EducationRepository) EducationService {
	return &educationService{educationRepo: educationRepo}
}

func (s *educationService) Create(input dto.CreateEducationDTO, createdBy uint, thumbnailPath string) (models.Education, error) {
	education := models.Education{
		Name:      input.Name,
		Url:       input.Url,
		Thumbnail: thumbnailPath, // Simpan path/URL dari handler
		CreatedBy: createdBy,
	}
	created, err := s.educationRepo.Create(education)
	if err != nil {
		return models.Education{}, fmt.Errorf("gagal membuat edukasi di db: %w", err)
	}
	return created, nil
}

func (s *educationService) FindAll() ([]models.Education, error) {
	return s.educationRepo.FindAll()
}

func (s *educationService) FindByID(id uint) (models.Education, error) {
	return s.educationRepo.FindByID(id)
}

func (s *educationService) Update(id uint, input dto.UpdateEducationDTO, thumbnailPath *string) (models.Education, error) {
	education, err := s.educationRepo.FindByID(id)
	if err != nil {
		return models.Education{}, fmt.Errorf("edukasi tidak ditemukan: %w", err)
	}

	oldThumbnailPath := education.Thumbnail // Simpan path lama
	shouldDeleteOld := false

	education.Name = input.Name
	education.Url = input.Url

	// Update thumbnail hanya jika path baru diberikan
	if thumbnailPath != nil {
		education.Thumbnail = *thumbnailPath
		// Tandai file lama untuk dihapus SETELAH update DB berhasil
		if oldThumbnailPath != "" && oldThumbnailPath != *thumbnailPath {
			shouldDeleteOld = true
		}
	}

	// Update data di database
	updatedEducation, err := s.educationRepo.Update(education)
	if err != nil {
		return models.Education{}, fmt.Errorf("gagal update edukasi di db: %w", err)
	}

	// Hapus file lama jika update DB berhasil dan path baru berbeda
	if shouldDeleteOld {
		go func(pathToDelete string) { // Hapus di background goroutine
			err := os.Remove(pathToDelete) // Perlu path absolut atau relatif dari CWD
			if err != nil {
				log.Printf("WARNING: Gagal menghapus thumbnail lama '%s': %v", pathToDelete, err)
			} else {
				log.Printf("Thumbnail lama '%s' berhasil dihapus.", pathToDelete)
			}
		}(oldThumbnailPath)
	}

	return updatedEducation, nil
}
func (s *educationService) Delete(id uint) error {
	education, err := s.educationRepo.FindByID(id)
	if err != nil {
		// Jika tidak ditemukan, anggap sudah terhapus
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("gagal mencari edukasi untuk dihapus: %w", err)
	}

	thumbnailPathToDelete := education.Thumbnail

	// Hapus data dari database
	err = s.educationRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus edukasi dari db: %w", err)
	}

	// Hapus file jika path ada dan delete DB berhasil
	if thumbnailPathToDelete != "" {
		go func(pathToDelete string) { // Hapus di background
			err := os.Remove(pathToDelete)
			if err != nil {
				log.Printf("WARNING: Gagal menghapus thumbnail '%s' saat delete edukasi: %v", pathToDelete, err)
			} else {
				log.Printf("Thumbnail '%s' berhasil dihapus saat delete edukasi.", pathToDelete)
			}
		}(thumbnailPathToDelete)
	}

	return nil
}
package services

import (
	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type EducationService interface {
	Create(input dto.CreateEducationDTO, createdBy uint) (models.Education, error)
	FindAll() ([]models.Education, error)
	FindByID(id uint) (models.Education, error)
	Update(id uint, input dto.UpdateEducationDTO) (models.Education, error)
	Delete(id uint) error
}

type educationService struct {
	educationRepo repositories.EducationRepository
}

func NewEducationService(educationRepo repositories.EducationRepository) EducationService {
	return &educationService{educationRepo: educationRepo}
}

func (s *educationService) Create(input dto.CreateEducationDTO, createdBy uint) (models.Education, error) {
	education := models.Education{
		Name:      input.Name,
		Url:       input.Url,
		Thumbnail: input.Thumbnail,
		CreatedBy: createdBy,
	}
	return s.educationRepo.Create(education)
}

func (s *educationService) FindAll() ([]models.Education, error) {
	return s.educationRepo.FindAll()
}

func (s *educationService) FindByID(id uint) (models.Education, error) {
	return s.educationRepo.FindByID(id)
}

func (s *educationService) Update(id uint, input dto.UpdateEducationDTO) (models.Education, error) {
	education, err := s.educationRepo.FindByID(id)
	if err != nil {
		return models.Education{}, err
	}
	education.Name = input.Name
	education.Url = input.Url
	education.Thumbnail = input.Thumbnail
	return s.educationRepo.Update(education)
}

func (s *educationService) Delete(id uint) error {
	return s.educationRepo.Delete(id)
}
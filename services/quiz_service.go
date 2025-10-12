package services

import (
	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type QuizService interface {
	Create(input dto.CreateQuizDTO, createdBy uint) (models.Quiz, error)
	FindAll() ([]models.Quiz, error)
	FindByID(id uint) (models.Quiz, error)
	Update(id uint, input dto.UpdateQuizDTO) (models.Quiz, error)
	Delete(id uint) error
}

type quizService struct {
	quizRepo repositories.QuizRepository
}

func NewQuizService(quizRepo repositories.QuizRepository) QuizService {
	return &quizService{quizRepo: quizRepo}
}

func (s *quizService) Create(input dto.CreateQuizDTO, createdBy uint) (models.Quiz, error) {
	quiz := models.Quiz{
		Name:      input.Name,
		Url:       input.Url,
		CreatedBy: createdBy,
	}
	return s.quizRepo.Create(quiz)
}

func (s *quizService) FindAll() ([]models.Quiz, error) {
	return s.quizRepo.FindAll()
}

func (s *quizService) FindByID(id uint) (models.Quiz, error) {
	return s.quizRepo.FindByID(id)
}

func (s *quizService) Update(id uint, input dto.UpdateQuizDTO) (models.Quiz, error) {
	quiz, err := s.quizRepo.FindByID(id)
	if err != nil {
		return models.Quiz{}, err
	}
	quiz.Name = input.Name
	quiz.Url = input.Url
	return s.quizRepo.Update(quiz)
}

func (s *quizService) Delete(id uint) error {
	return s.quizRepo.Delete(id)
}
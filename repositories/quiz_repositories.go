package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type QuizRepository interface {
	Create(quiz models.Quiz) (models.Quiz, error)
	FindAll() ([]models.Quiz, error)
	FindByID(id uint) (models.Quiz, error)
	Update(quiz models.Quiz) (models.Quiz, error)
	Delete(id uint) error
}

type quizRepository struct {
	db *gorm.DB
}

func NewQuizRepository(db *gorm.DB) QuizRepository {
	return &quizRepository{db: db}
}

func (r *quizRepository) Create(quiz models.Quiz) (models.Quiz, error) {
	err := r.db.Create(&quiz).Error
	return quiz, err
}

func (r *quizRepository) FindAll() ([]models.Quiz, error) {
	var quizzes []models.Quiz
	err := r.db.Find(&quizzes).Error
	return quizzes, err
}

func (r *quizRepository) FindByID(id uint) (models.Quiz, error) {
	var quiz models.Quiz
	err := r.db.First(&quiz, id).Error
	return quiz, err
}

func (r *quizRepository) Update(quiz models.Quiz) (models.Quiz, error) {
	err := r.db.Save(&quiz).Error
	return quiz, err
}

func (r *quizRepository) Delete(id uint) error {
	return r.db.Delete(&models.Quiz{}, id).Error
}
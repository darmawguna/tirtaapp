package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

// UserRepository mendefinisikan interface untuk operasi data user.
type UserRepository interface {
	CreateUser(user models.User) (models.User, error)
	FindByEmail(email string) (models.User, error) 
	FindByID(id uint) (models.User, error) 
	 // <-- Tambahkan/Pastikan ada
	Update(user models.User) (models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru dari userRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// CreateUser menyimpan user baru ke database.
func (r *userRepository) CreateUser(user models.User) (models.User, error) {
	err := r.db.Create(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *userRepository) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return models.User{}, err // GORM akan return ErrRecordNotFound jika tidak ada
	}
	return user, nil
}

func (r *userRepository) FindByID(id uint) (models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return models.User{}, err // GORM akan return ErrRecordNotFound jika tidak ada
	}
	return user, nil
}

func (r *userRepository) Update(user models.User) (models.User, error) {
	// GORM Save akan memperbarui semua field record 'user' berdasarkan Primary Key nya.
	err := r.db.Save(&user).Error
	return user, err
}
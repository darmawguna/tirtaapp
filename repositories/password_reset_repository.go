package repositories

import (
	"time"

	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type PasswordResetRepository interface {
	Create(reset models.PasswordReset) (models.PasswordReset, error)
	FindValidByTokenHash(tokenHash string, now time.Time) (models.PasswordReset, error)
	MarkUsed(id uint, usedAt time.Time) error
	DeleteActiveByUser(userID uint) error
}

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(reset models.PasswordReset) (models.PasswordReset, error) {
	err := r.db.Create(&reset).Error
	return reset, err
}

func (r *passwordResetRepository) FindValidByTokenHash(tokenHash string, now time.Time) (models.PasswordReset, error) {
	var pr models.PasswordReset
	err := r.db.
		Where("token_hash = ?", tokenHash).
		Where("used_at IS NULL").
		Where("expires_at > ?", now).
		First(&pr).Error
	return pr, err
}

func (r *passwordResetRepository) MarkUsed(id uint, usedAt time.Time) error {
	return r.db.Model(&models.PasswordReset{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", usedAt).Error
}

// Supaya 1 user gak punya banyak token aktif (anti spam, enak buat demo)
func (r *passwordResetRepository) DeleteActiveByUser(userID uint) error {
	return r.db.
		Where("user_id = ? AND used_at IS NULL", userID).
		Delete(&models.PasswordReset{}).Error
}

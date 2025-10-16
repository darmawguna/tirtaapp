package repositories

import (
	models "github.com/darmawguna/tirtaapp.git/model"
	"gorm.io/gorm"
)

type DeviceRepository interface {
	FindByToken(token string) (models.Device, error)
	CreateOrUpdate(device models.Device) (models.Device, error)
	FindAllByUserID(userID uint) ([]models.Device, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

// FindByToken mencari perangkat berdasarkan FCM token.
func (r *deviceRepository) FindByToken(token string) (models.Device, error) {
	var device models.Device
	err := r.db.Where("fcm_token = ?", token).First(&device).Error
	return device, err
}

// CreateOrUpdate membuat record baru jika token tidak ada, atau memperbarui jika sudah ada.
func (r *deviceRepository) CreateOrUpdate(device models.Device) (models.Device, error) {
	// GORM's Save akan otomatis melakukan INSERT atau UPDATE berdasarkan Primary Key.
	// Jika device.ID adalah 0, ia akan INSERT. Jika tidak, ia akan UPDATE.
	err := r.db.Save(&device).Error
	return device, err
}

func (r *deviceRepository) FindAllByUserID(userID uint) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}
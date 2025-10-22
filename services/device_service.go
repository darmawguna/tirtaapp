package services

import (
	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm"
)

type DeviceService interface {
	RegisterDevice(userID uint, input dto.RegisterDeviceDTO) (models.Device, error)
}

type deviceService struct {
	deviceRepo repositories.DeviceRepository
}

func NewDeviceService(deviceRepo repositories.DeviceRepository) DeviceService {
	return &deviceService{deviceRepo: deviceRepo}
}

func (s *deviceService) RegisterDevice(userID uint, input dto.RegisterDeviceDTO) (models.Device, error) {
	// Cek apakah token sudah ada di database
	existingDevice, err := s.deviceRepo.FindByToken(input.FCMToken)

	// Jika tidak ada error selain 'record not found', berarti ada masalah lain
	if err != nil && err != gorm.ErrRecordNotFound {
		return models.Device{}, err
	}

	// Jika token sudah ada (existingDevice.ID != 0)
	if existingDevice.ID != 0 {
		// Update UserID jika token tersebut sekarang digunakan oleh user lain
		existingDevice.UserID = userID
		existingDevice.DeviceType = input.DeviceType
		return s.deviceRepo.CreateOrUpdate(existingDevice)
	}

	// Jika token belum ada, buat record baru
	newDevice := models.Device{
		UserID:     userID,
		FCMToken:   input.FCMToken,
		DeviceType: input.DeviceType,
	}
	return s.deviceRepo.CreateOrUpdate(newDevice)
}
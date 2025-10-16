package dto

// RegisterDeviceDTO adalah DTO untuk request registrasi token perangkat.
type RegisterDeviceDTO struct {
	FCMToken   string `json:"fcm_token" binding:"required"`
	DeviceType string `json:"device_type" binding:"omitempty,oneof=android ios"` // Opsional
}
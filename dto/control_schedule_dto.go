package dto

// CreateControlScheduleDTO adalah DTO untuk membuat jadwal kontrol baru.
type CreateControlScheduleDTO struct {
	ControlDate string `json:"control_date" binding:"required,datetime=2006-01-02"`
	Notes       string `json:"notes"`
}

// UpdateControlScheduleDTO adalah DTO untuk memperbarui jadwal kontrol.
type UpdateControlScheduleDTO struct {
	ControlDate string `json:"control_date" binding:"required,datetime=2006-01-02"`
	Notes       string `json:"notes"`
	IsActive    *bool  `json:"is_active" binding:"required"`
}

type ControlScheduleResponseDTO struct {
	ID          uint   `json:"id"`
	UserID      uint   `json:"user_id"`
	ControlDate string `json:"control_date"`
	Notes       string `json:"notes"`
	IsActive    bool   `json:"is_active"`
}
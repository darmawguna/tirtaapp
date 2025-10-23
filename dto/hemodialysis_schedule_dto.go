package dto

// CreateHemodialysisScheduleDTO adalah DTO untuk membuat jadwal hemodialisa baru.
type CreateHemodialysisScheduleDTO struct {
	ScheduleDate string `json:"schedule_date" binding:"required,datetime=2006-01-02"`
}

// UpdateHemodialysisScheduleDTO adalah DTO untuk memperbarui jadwal hemodialisa.
type UpdateHemodialysisScheduleDTO struct {
	ScheduleDate string `json:"schedule_date" binding:"required,datetime=2006-01-02"`
	IsActive     *bool  `json:"is_active" binding:"required"`
}

type HemodialysisScheduleResponseDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	ScheduleDate string `json:"schedule_date"`
	IsActive     bool   `json:"is_active"`
}
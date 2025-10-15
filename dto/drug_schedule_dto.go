package dto

// CreateDrugScheduleDTO adalah DTO untuk membuat jadwal minum obat baru.
type CreateDrugScheduleDTO struct {
	DrugName     string `json:"drug_name" binding:"required"`
	Dose         string `json:"dose" binding:"required"`
	ScheduleDate string `json:"schedule_date" binding:"required,datetime=2006-01-02"`
	At06         bool   `json:"at_06"`
	At12         bool   `json:"at_12"`
	At18         bool   `json:"at_18"`
}

// UpdateDrugScheduleDTO adalah DTO untuk memperbarui jadwal minum obat.
type UpdateDrugScheduleDTO struct {
	DrugName     string `json:"drug_name" binding:"required"`
	Dose         string `json:"dose" binding:"required"`
	ScheduleDate string `json:"schedule_date" binding:"required,datetime=2006-01-02"`
	At06         bool   `json:"at_06"`
	At12         bool   `json:"at_12"`
	At18         bool   `json:"at_18"`
	IsActive     *bool  `json:"is_active" binding:"required"`
}

type DrugScheduleResponseDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	DrugName     string `json:"drug_name"`
	Dose         string `json:"dose"`
	ScheduleDate string `json:"schedule_date"`
	At06         bool   `json:"at_06"`
	At12         bool   `json:"at_12"`
	At18         bool   `json:"at_18"`
	IsActive     bool   `json:"is_active"`
}
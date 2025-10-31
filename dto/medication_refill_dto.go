package dto

// DTO untuk input (Create/Update)
type CreateMedicationRefillDTO struct {
	RefillDate   string `json:"refill_date" binding:"required,datetime=2006-01-02"`
}

type UpdateMedicationRefillDTO struct {
	RefillDate   string `json:"refill_date" binding:"required,datetime=2006-01-02"`
	IsActive     *bool  `json:"is_active" binding:"required"`
}

// DTO untuk response
type MedicationRefillResponseDTO struct {
	ID           uint   `json:"id"`
	UserID       uint   `json:"user_id"`
	RefillDate   string `json:"refill_date"`
	IsActive     bool   `json:"is_active"`
}
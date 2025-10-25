package dto

type CreateOrUpdateFluidLogDTO struct {
	IntakeCC *int `json:"intake_cc" binding:"required,min=0"`
	OutputCC *int `json:"output_cc" binding:"required,min=0"`
}

type FluidBalanceLogResponseDTO struct {
	ID             uint   `json:"id"`
	UserID         uint   `json:"user_id"`
	LogDate        string `json:"log_date"` // Format YYYY-MM-DD
	IntakeCC       int    `json:"intake_cc"`
	OutputCC       int    `json:"output_cc"`
	BalanceCC      int    `json:"balance_cc"`
	WarningMessage string `json:"warning_message,omitempty"` // Hanya muncul jika ada warning
}
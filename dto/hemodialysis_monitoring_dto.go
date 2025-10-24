package dto

type CreateHemodialysisMonitoringDTO struct {
	BPBefore     string  `json:"bp_before" binding:"required"`
	BPAfter      string  `json:"bp_after" binding:"required"`
	WeightBefore float64 `json:"weight_before" binding:"required,gt=0"`
	WeightAfter  float64 `json:"weight_after" binding:"required,gt=0"`
}

type HemodialysisMonitoringResponseDTO struct {
	ID             uint    `json:"id"`
	UserID         uint    `json:"user_id"`
	MonitoringDate string  `json:"monitoring_date"` // Format YYYY-MM-DD
	BPBefore       string  `json:"bp_before"`
	BPAfter        string  `json:"bp_after"`
	WeightBefore   float64 `json:"weight_before"`
	WeightAfter    float64 `json:"weight_after"`
}
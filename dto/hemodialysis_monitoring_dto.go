package dto

type CreateHemodialysisMonitoringDTO struct {
	BPBefore     string  `json:"bp_before" binding:"required"`
	BPAfter      string  `json:"bp_after" binding:"required"`
	WeightBefore float64 `json:"weight_before" binding:"required,gt=0"`
	WeightAfter  float64 `json:"weight_after" binding:"required,gt=0"`
}

type HemodialysisMonitoringResponseDTO struct {
	ID                     uint    `json:"id"`
	UserID                 uint    `json:"user_id"`
	HemodialysisScheduleID uint    `json:"hemodialysis_schedule_id"`
	ScheduleDate           string  `json:"schedule_date"` // Get date from the linked schedule
	BPBefore               string  `json:"bp_before"`
	BPAfter                string  `json:"bp_after"`
	WeightBefore           float64 `json:"weight_before"`
	WeightAfter            float64 `json:"weight_after"`
}
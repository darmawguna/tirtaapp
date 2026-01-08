package dto

import "time"

type ComplaintPhase string

const (
	PhasePreHD  ComplaintPhase = "pre_hd"
	PhasePostHD ComplaintPhase = "post_hd"
)

type CreateComplaintDTO struct {
	Phase     ComplaintPhase `json:"phase" binding:"required"`
	Complaints []string      `json:"complaints" binding:"required,min=1"`
	OtherText *string        `json:"other_text"`
}

type ComplaintItem struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

type ComplaintLogResponse struct {
	ID              uint           `json:"id"`
	Phase           ComplaintPhase `json:"phase"`
	Complaints      []ComplaintItem `json:"complaints"`
	OtherText       *string        `json:"other_text"`
	GeneratedMessage string        `json:"generated_message"`
	CreatedAt       time.Time      `json:"created_at"`
}

package dto

// CreateComplaintDTO adalah DTO untuk request pembuatan keluhan.
type CreateComplaintDTO struct {
	Complaints []string `json:"complaints" binding:"required,min=1"`
}

type ComplaintResponse struct {
	ID        string   `json:"id" binding:"required,url"`
	Complaint []string `json:"complaints"`
	Message   string   `json:"message" binding:"required,url"`
}

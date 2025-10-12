package dto

// CreateQuizDTO adalah DTO untuk membuat kuis baru.
type CreateQuizDTO struct {
	Name string `json:"name" binding:"required"`
	Url  string `json:"url" binding:"required,url"`
}

// UpdateQuizDTO adalah DTO untuk memperbarui kuis.
type UpdateQuizDTO struct {
	Name string `json:"name" binding:"required"`
	Url  string `json:"url" binding:"required,url"`
}

// TODO buat api response general untuk semua response api

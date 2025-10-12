package dto

// CreateEducationDTO adalah DTO untuk membuat data edukasi baru.
type CreateEducationDTO struct {
	Name      string `json:"name" binding:"required"`
	Url       string `json:"url" binding:"required,url"`
	Thumbnail string `json:"thumbnail" binding:"required,url"`
}

// UpdateEducationDTO adalah DTO untuk memperbarui data edukasi.
type UpdateEducationDTO struct {
	Name      string `json:"name" binding:"required"`
	Url       string `json:"url" binding:"required,url"`
	Thumbnail string `json:"thumbnail" binding:"required,url"`
}
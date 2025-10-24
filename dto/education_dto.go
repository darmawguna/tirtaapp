package dto

// CreateEducationDTO adalah DTO untuk membuat data edukasi baru.
type CreateEducationDTO struct {
	Name string `json:"name" form:"name"` // Hapus binding:"required"
	Url  string `json:"url" form:"url"`   // Hapus binding:"required,url"
}

// UpdateEducationDTO (Hapus tag binding)
type UpdateEducationDTO struct {
	Name string `json:"name" form:"name"` // Hapus binding:"required"
	Url  string `json:"url" form:"url"`   // Hapus binding:"required,url"
}

type EducationResponseDTO struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Thumbnail string `json:"thumbnail"` // Tetap ada untuk response
	CreatedBy uint   `json:"created_by"`
}
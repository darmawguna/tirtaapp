package dto

// UserResponseDTO adalah DTO untuk response data user.
type UserResponseDTO struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	PhoneNumber string `json:"phone_number"`
	ProfilePicture string `json:"profile_picture,omitempty"`
}

type UpdateProfileDTO struct {
	Name     *string `json:"name" form:"name" binding:"omitempty"`
	Password *string `json:"password" form:"password" binding:"omitempty,min=6"`
}
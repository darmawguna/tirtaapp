package dto

// RegisterDTO adalah DTO untuk request registrasi.
type RegisterDTO struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role"`
	Password string `json:"password" binding:"required,min=6"`
	Timezone string `json:"timezone" binding:"required"`
}
 
type LoginDTO struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	FCMToken string `json:"fcm_token" binding:"required"`
}
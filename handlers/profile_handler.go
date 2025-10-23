package handlers

import (
	"net/http"
	"strings" // Import strings

	"github.com/darmawguna/tirtaapp.git/dto" // Sesuaikan path
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

// ProfileHandler mengelola request terkait profil user.
type ProfileHandler struct {
	profileService services.ProfileService
}

// NewProfileHandler adalah constructor untuk ProfileHandler.
func NewProfileHandler(profileService services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// toUserResponseDTO mengonversi model User ke DTO response (menghilangkan password).
func toUserResponseDTO(user models.User) dto.UserResponseDTO {
	return dto.UserResponseDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		// Timezone bisa ditambahkan jika perlu ditampilkan
	}
}

// GetProfile menangani permintaan GET /api/v1/profile.
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// Ambil userID dari context (ditetapkan oleh AuthMiddleware)
	userID := c.MustGet("userID").(float64) // Tipe default dari JWT claim

	// Panggil service untuk mendapatkan profil
	user, err := h.profileService.GetProfile(uint(userID))
	if err != nil {
		// Jika service mengembalikan error "not found"
		c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
		return
	}

	// Konversi ke DTO response dan kirim
	responseDTO := toUserResponseDTO(user)
	c.JSON(http.StatusOK, utils.SuccessResponse("Profile fetched successfully", responseDTO))
}

// UpdateProfile menangani permintaan PUT /api/v1/profile.
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	var input dto.UpdateProfileDTO
	// Bind JSON dan jalankan validasi (omitempty, min=6 untuk password)
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	// Panggil service untuk update profil
	updatedUser, err := h.profileService.UpdateProfile(uint(userID), input)
	if err != nil {
		// Tangani kemungkinan error dari service
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update profile", err.Error()))
		return
	}

	// Konversi ke DTO response dan kirim
	responseDTO := toUserResponseDTO(updatedUser)
	c.JSON(http.StatusOK, utils.SuccessResponse("Profile updated successfully", responseDTO))
}
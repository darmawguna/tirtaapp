package handlers

import (
	"fmt"            // Import fmt
	"log"            // Import log
	"mime/multipart" // Import multipart
	"net/http"
	"os"            // Import os
	"path/filepath" // Import filepath
	"strings"
	"time" // Import time

	"github.com/darmawguna/tirtaapp.git/dto"          // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper" // Import viper
)

const UPLOAD_PATH_PROFILE = "./uploads/profiles" // Definisikan path upload profile

// ProfileHandler mengelola request terkait profil user.
type ProfileHandler struct {
	profileService services.ProfileService
}

// NewProfileHandler adalah constructor untuk ProfileHandler.
func NewProfileHandler(profileService services.ProfileService) *ProfileHandler {
	// Pastikan direktori upload ada
	if err := os.MkdirAll(UPLOAD_PATH_PROFILE, os.ModePerm); err != nil {
		log.Fatalf("FATAL: Tidak bisa membuat direktori upload profile: %v", err)
	}
	return &ProfileHandler{profileService: profileService}
}

// toUserResponseDTO (Diperbarui)
func toUserResponseDTO(user models.User) dto.UserResponseDTO {
	profilePictureUrl := ""
	if user.ProfilePicture != "" {
		// Asumsi path relatif disimpan di DB, buat URL lengkap
		baseUrl := viper.GetString("BASE_URL")
		if baseUrl == "" { baseUrl = "http://localhost:8080" } // Fallback
		// Sesuaikan "/static/profiles/" jika path serving berbeda
		profilePictureUrl = fmt.Sprintf("%s/static/profiles/%s", baseUrl, filepath.Base(user.ProfilePicture))
	}
	return dto.UserResponseDTO{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		PhoneNumber: user.PhoneNumber,
		ProfilePicture: profilePictureUrl, // <-- Kirim URL lengkap
		Role:           user.Role,
	}
}

// GetProfile (Tidak berubah)
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	user, err := h.profileService.GetProfile(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
		return
	}
	responseDTO := toUserResponseDTO(user)
	c.JSON(http.StatusOK, utils.SuccessResponse("Profile fetched successfully", responseDTO))
}

// UpdateProfile (Diperbarui untuk multipart/form-data)
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	var input dto.UpdateProfileDTO
	// [PEMBARUAN] Bind dari form-data (Name, Password)
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed for Name/Password", err.Error()))
		return
	}

	var newProfilePicturePath *string // Pointer untuk path file baru
	file, err := c.FormFile("profile_picture") // Nama field file di form-data

	// Jika ada file baru di request
	if err == nil {
		// Validasi file (contoh: ukuran < 2MB)
		if file.Size > 5*1024*1024 { // 2 MB
			c.JSON(http.StatusBadRequest, utils.ErrorResponse("Profile picture size exceeds 2MB limit", nil))
			return
		}
		// Anda bisa menambahkan validasi tipe file (misal: image/jpeg, image/png)

		// Simpan file baru
		savedPath, saveErr := saveUploadedProfileFile(c, file, UPLOAD_PATH_PROFILE) // Gunakan helper baru/spesifik
		if saveErr != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to save profile picture", saveErr.Error()))
			return
		}
		newProfilePicturePath = &savedPath // Set pointer ke path baru
	} else if err != http.ErrMissingFile {
		// Error selain karena file tidak ada
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Error processing profile picture file", err.Error()))
		return
	}
	// Jika err == http.ErrMissingFile, tidak ada file baru, newProfilePicturePath tetap nil

	// Panggil service update dengan DTO dan path file baru (atau nil)
	updatedUser, err := h.profileService.UpdateProfile(uint(userID), input, newProfilePicturePath)
	if err != nil {
		// Tangani error dari service
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		// Jika gagal update DB, dan ada file baru yang tersimpan, coba hapus lagi
		if newProfilePicturePath != nil {
			go os.Remove(*newProfilePicturePath)
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update profile", err.Error()))
		return
	}

	// Konversi ke DTO response dan kirim
	responseDTO := toUserResponseDTO(updatedUser)
	c.JSON(http.StatusOK, utils.SuccessResponse("Profile updated successfully", responseDTO))
}


// --- Helper Function untuk Menyimpan File Foto Profil ---
// Mirip dengan saveUploadedFile di education handler, tapi path berbeda
// Mengembalikan path absolut file yang disimpan
func saveUploadedProfileFile(c *gin.Context, file *multipart.FileHeader, destDirectory string) (string, error) {
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.Filename, ext)
	safeBaseName := strings.ReplaceAll(baseName, " ", "_")
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), safeBaseName, ext)

	dst := filepath.Join(destDirectory, filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", fmt.Errorf("gagal menyimpan file ke '%s': %w", dst, err)
	}
	log.Printf("File '%s' berhasil disimpan sebagai '%s' di: %s", file.Filename, filename, dst)

	return dst, nil
}
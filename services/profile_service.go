package services

import (
	"errors" // Tambahkan import errors
	"fmt"
	"log"
	"os"

	"github.com/darmawguna/tirtaapp.git/dto" // Sesuaikan path module Anda
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"golang.org/x/crypto/bcrypt"
)

// ProfileService mendefinisikan interface untuk operasi profil.
type ProfileService interface {
	GetProfile(userID uint) (models.User, error)
	UpdateProfile(userID uint, input dto.UpdateProfileDTO, profilePicturePath *string) (models.User, error)
}

// profileService adalah implementasi dari ProfileService.
type profileService struct {
	userRepo repositories.UserRepository
}

// NewProfileService adalah constructor untuk profileService.
func NewProfileService(userRepo repositories.UserRepository) ProfileService {
	return &profileService{userRepo: userRepo}
}

// GetProfile mengambil data profil pengguna berdasarkan ID.
func (s *profileService) GetProfile(userID uint) (models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		// Kembalikan error yang lebih spesifik jika user tidak ditemukan
		return models.User{}, errors.New("user profile not found")
	}
	return user, nil
}

// UpdateProfile memperbarui data profil pengguna (nama dan/atau password).
func (s *profileService) UpdateProfile(userID uint, input dto.UpdateProfileDTO, profilePicturePath *string) (models.User, error) {
	// 1. Ambil data user saat ini
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return models.User{}, errors.New("user profile not found")
	}

	oldProfilePicturePath := user.ProfilePicture // Simpan path lama
	shouldDeleteOld := false

	// 2. Perbarui nama jika ada di input
	if input.Name != nil && *input.Name != "" {
		user.Name = *input.Name
	}

	// 3. Perbarui password jika ada di input
	if input.Password != nil && *input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, fmt.Errorf("failed to hash new password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	// 4. [PEMBARUAN] Perbarui foto profil jika path baru diberikan
	if profilePicturePath != nil {
		user.ProfilePicture = *profilePicturePath // Simpan path file baru
		// Tandai file lama untuk dihapus jika pathnya ada dan berbeda
		if oldProfilePicturePath != "" && oldProfilePicturePath != *profilePicturePath {
			shouldDeleteOld = true
		}
	}

	// 5. Simpan semua perubahan ke database
	updatedUser, err := s.userRepo.Update(user)
	if err != nil {
		// Jika update DB gagal, jangan hapus file lama
		return models.User{}, fmt.Errorf("failed to update profile in database: %w", err)
	}

	// 6. [PEMBARUAN] Hapus file lama jika update DB berhasil
	if shouldDeleteOld {
		go func(pathToDelete string) { // Hapus di background
			log.Printf("Attempting to delete old profile picture: %s", pathToDelete)
			err := os.Remove(pathToDelete) // Pastikan path ini absolut atau relatif dari CWD API
			if err != nil {
				log.Printf("WARNING: Failed to delete old profile picture '%s': %v", pathToDelete, err)
			} else {
				log.Printf("Old profile picture '%s' deleted successfully.", pathToDelete)
			}
		}(oldProfilePicturePath)
	}

	return updatedUser, nil
}
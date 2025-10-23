package services

import (
	"errors" // Tambahkan import errors
	"fmt"

	"github.com/darmawguna/tirtaapp.git/dto" // Sesuaikan path module Anda
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"golang.org/x/crypto/bcrypt"
)

// ProfileService mendefinisikan interface untuk operasi profil.
type ProfileService interface {
	GetProfile(userID uint) (models.User, error)
	UpdateProfile(userID uint, input dto.UpdateProfileDTO) (models.User, error)
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
func (s *profileService) UpdateProfile(userID uint, input dto.UpdateProfileDTO) (models.User, error) {
	// 1. Ambil data user saat ini
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return models.User{}, errors.New("user profile not found")
	}

	// 2. Perbarui nama jika ada di input
	if input.Name != nil && *input.Name != "" { // Pastikan tidak string kosong juga
		user.Name = *input.Name
	}

	// 3. Perbarui password jika ada di input
	if input.Password != nil && *input.Password != "" { // Pastikan tidak string kosong juga
		// Hash password baru
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return models.User{}, fmt.Errorf("failed to hash new password: %w", err)
		}
		user.Password = string(hashedPassword)
	}

	// 4. Simpan perubahan ke database
	updatedUser, err := s.userRepo.Update(user)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to update profile in database: %w", err)
	}

	return updatedUser, nil
}
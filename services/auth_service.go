package services

import (
	"errors"
	"log"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"github.com/darmawguna/tirtaapp.git/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(input dto.RegisterDTO) (models.User, error)
	Login(input dto.LoginDTO) (string, error)
}

type authService struct {
	userRepository repositories.UserRepository
	deviceService    DeviceService
}

// NewAuthService membuat instance baru dari authService.
func NewAuthService(userRepo repositories.UserRepository, deviceService DeviceService) AuthService {
	return &authService{userRepository: userRepo, deviceService: deviceService}
}

func (s *authService) Register(input dto.RegisterDTO) (models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	// Buat objek user baru
	newUser := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role, // Default role
		Timezone: input.Timezone,
	}

	// Simpan user ke database via repository
	createdUser, err := s.userRepository.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}

	return createdUser, nil
}

func (s *authService) Login(input dto.LoginDTO) (string, error) {
	// 1. Cari user berdasarkan email
	user, err := s.userRepository.FindByEmail(input.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	// 2. Bandingkan password yang di-hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		// Jika error (password tidak cocok)
		return "", errors.New("invalid email or password")
	}
	
	deviceDTO := dto.RegisterDeviceDTO{FCMToken: input.FCMToken}
	_, err = s.deviceService.RegisterDevice(user.ID, deviceDTO)
	if err != nil {
		// Log error ini, tapi jangan sampai menggagalkan login
		log.Printf("WARNING: Failed to register device for user %d: %v", user.ID, err)
	}

	// 3. Jika berhasil, generate JWT
	token, err := utils.GenerateJWT(user)
	if err != nil {
		return "", err
	}

	return token, nil
}


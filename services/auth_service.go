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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}
	newUser := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
		Role:     input.Role, // Default role
		Timezone: input.Timezone,
		PhoneNumber: input.PhoneNumber,
	}
	createdUser, err := s.userRepository.CreateUser(newUser)
	if err != nil {
		return models.User{}, err
	}
	return createdUser, nil
}

func (s *authService) Login(input dto.LoginDTO) (string, error) {
	user, err := s.userRepository.FindByEmail(input.Email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	if input.Timezone != "" && user.Timezone != input.Timezone {
		log.Printf("Updating timezone for user %d from '%s' to '%s'", user.ID, user.Timezone, input.Timezone)
		user.Timezone = input.Timezone
		_, updateErr := s.userRepository.UpdateTimeZone(user)
		if updateErr != nil {
			log.Printf("WARNING: Failed to update timezone for user %d: %v", user.ID, updateErr)
		} else {
			log.Printf("Timezone updated successfully for user %d", user.ID)
		}
	}
	deviceDTO := dto.RegisterDeviceDTO{FCMToken: input.FCMToken}
	_, err = s.deviceService.RegisterDevice(user.ID, deviceDTO)
	if err != nil {
		log.Printf("WARNING: Failed to register device for user %d: %v", user.ID, err)
	}
	token, err := utils.GenerateJWT(user)
	if err != nil {
		return "", err
	}
	return token, nil
}


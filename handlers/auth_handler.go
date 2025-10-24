package handlers

import (
	"net/http"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var input dto.RegisterDTO

	// Binding dan validasi request body ke DTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Validation failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Panggil service untuk proses registrasi
	user, err := h.authService.Register(input)
	if err != nil {
		// Nanti kita bisa buat error handling yang lebih baik
		response := utils.ErrorResponse("Registration failed", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}
	userResponse := dto.UserResponseDTO{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		PhoneNumber: user.PhoneNumber,
	}
	response := utils.SuccessResponse("User registered successfully", userResponse)
	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginDTO

	// Binding dan validasi request body
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Login failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Panggil service login
	token, err := h.authService.Login(input)
	if err != nil {
		// Cek apakah error karena kredensial tidak valid
		if strings.Contains(err.Error(), "invalid email or password") {
			response := utils.ErrorResponse("Login failed", err.Error())
			c.JSON(http.StatusBadRequest, response)
			return
		}
		// Error server lainnya
		response := utils.ErrorResponse("Login failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Kirim response sukses dengan token
	c.JSON(http.StatusOK, gin.H{"token": token})
}

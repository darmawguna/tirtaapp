package handlers

import (
	"net/http"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Panggil service untuk proses registrasi
	user, err := h.authService.Register(input)
	if err != nil {
		// Nanti kita bisa buat error handling yang lebih baik
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Kirim response sukses
	// Kita tidak mengirim password kembali dalam response
	response := gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"role":  user.Role,
	}
	c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input dto.LoginDTO

	// Binding dan validasi request body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Panggil service login
	token, err := h.authService.Login(input)
	if err != nil {
		// Cek apakah error karena kredensial tidak valid
		if strings.Contains(err.Error(), "invalid email or password") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Error server lainnya
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Kirim response sukses dengan token
	c.JSON(http.StatusOK, gin.H{"token": token})
}

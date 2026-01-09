package handlers

import (
	"net/http"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type PasswordResetHandler struct {
	service services.PasswordResetService
}

func NewPasswordResetHandler(service services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{service: service}
}

func (h *PasswordResetHandler) ForgotPassword(c *gin.Context) {
	var input dto.ForgotPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	email := strings.TrimSpace(input.Email)
	if err := h.service.ForgotPassword(email); err != nil {
		// internal error beneran -> 500, biar kamu gampang debug
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to process request", err.Error()))
		return
	}

	// Anti-enumeration: selalu 200 generik
	c.JSON(http.StatusOK, utils.SuccessResponse("If the email is registered, a reset link has been sent.", nil))
}

func (h *PasswordResetHandler) ResetPassword(c *gin.Context) {
	var input dto.ResetPasswordDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	if err := h.service.ResetPassword(input.Token, input.Password); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Reset password failed", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Password has been reset successfully", nil))
}

package handlers

import (
	"net/http"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type DeviceHandler struct {
	deviceService services.DeviceService
}

func NewDeviceHandler(deviceService services.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var input dto.RegisterDeviceDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)

	_, err := h.deviceService.RegisterDevice(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to register device", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Device registered successfully", nil))
}
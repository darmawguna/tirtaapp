package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type ControlScheduleHandler struct {
	service services.ControlScheduleService
}

func NewControlScheduleHandler(service services.ControlScheduleService) *ControlScheduleHandler {
	return &ControlScheduleHandler{service: service}
}

func toControlScheduleResponse(schedule models.ControlSchedule) dto.ControlScheduleResponseDTO {
	return dto.ControlScheduleResponseDTO{
		ID:          schedule.ID,
		UserID:      schedule.UserID,
		ControlDate: schedule.ControlDate.Format("2006-01-02"),
		Notes:       schedule.Notes,
		IsActive:    schedule.IsActive,
	}
}

func (h *ControlScheduleHandler) Create(c *gin.Context) {
	var input dto.CreateControlScheduleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Create(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create schedule", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse("Schedule created successfully", toControlScheduleResponse(schedule)))
}

func (h *ControlScheduleHandler) GetAll(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	schedules, err := h.service.FindAllByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch schedules", err.Error()))
		return
	}

	var responseDTOs []dto.ControlScheduleResponseDTO
	for _, s := range schedules {
		responseDTOs = append(responseDTOs, toControlScheduleResponse(s))
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Schedules fetched successfully", responseDTOs))
}

func (h *ControlScheduleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	schedule, err := h.service.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Drug schedule not found", err.Error()))
		return
	}
	
	// Verifikasi otorisasi
	userID := c.MustGet("userID").(float64)
	if schedule.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("You are not authorized to view this schedule", nil))
		return
	}

	response := utils.SuccessResponse("Control schedule fetched successfully", toControlScheduleResponse(schedule))
	c.JSON(http.StatusOK, response)
}

func (h *ControlScheduleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.MustGet("userID").(float64)

	var input dto.UpdateControlScheduleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	schedule, err := h.service.Update(uint(id), uint(userID), input)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update schedule", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Schedule updated successfully", toControlScheduleResponse(schedule)))
}

// Implementasikan GetByID dan Delete dengan pola yang sama...
func (h *ControlScheduleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	err = h.service.Delete(uint(id), uint(userID)) // Kirim userID ke service
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete drug schedule", err.Error()))
		return
	}

	response := utils.SuccessResponse("Control schedule deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
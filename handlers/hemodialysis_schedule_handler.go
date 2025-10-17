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

type HemodialysisScheduleHandler struct {
	service services.HemodialysisScheduleService
}

func NewHemodialysisScheduleHandler(service services.HemodialysisScheduleService) *HemodialysisScheduleHandler {
	return &HemodialysisScheduleHandler{service: service}
}

func toHemodialysisScheduleResponse(schedule models.HemodialysisSchedule) dto.HemodialysisScheduleResponseDTO {
	return dto.HemodialysisScheduleResponseDTO{
		ID:           schedule.ID,
		UserID:       schedule.UserID,
		ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
		Notes:        schedule.Notes,
		IsActive:     schedule.IsActive,
	}
}

func (h *HemodialysisScheduleHandler) GetAll(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	schedules, err := h.service.FindAllByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch schedules", err.Error()))
		return
	}

	var responseDTOs []dto.HemodialysisScheduleResponseDTO
	for _, s := range schedules {
		responseDTOs = append(responseDTOs, toHemodialysisScheduleResponse(s))
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Schedules fetched successfully", responseDTOs))
}

func (h *HemodialysisScheduleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	schedule, err := h.service.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Hemodialysis schedule not found", err.Error()))
		return
	}
	
	// Verifikasi otorisasi
	userID := c.MustGet("userID").(float64)
	if schedule.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("You are not authorized to view this Hemodialysis", nil))
		return
	}

	response := utils.SuccessResponse("Hemodialysis schedule fetched successfully", toHemodialysisScheduleResponse(schedule))
	c.JSON(http.StatusOK, response)
}

func (h *HemodialysisScheduleHandler) Create(c *gin.Context) {
	var input dto.CreateHemodialysisScheduleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Create(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create Hemodialysis", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse("Hemodialysis created successfully", toHemodialysisScheduleResponse(schedule)))
}

func (h *HemodialysisScheduleHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	userID := c.MustGet("userID").(float64)

	var input dto.UpdateHemodialysisScheduleDTO
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
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update Hemodialysis", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Hemodialysis updated successfully", toHemodialysisScheduleResponse(schedule)))
}

func (h *HemodialysisScheduleHandler) Delete(c *gin.Context) {
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
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete Hemodialysis schedule", err.Error()))
		return
	}

	response := utils.SuccessResponse("Hemodialysis schedule deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}

// Implementasikan GetAll, GetByID, Update, dan Delete dengan pola yang sama seperti ControlScheduleHandler
// ...
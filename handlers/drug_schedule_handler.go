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

type DrugScheduleHandler struct {
	service services.DrugScheduleService
}

func NewDrugScheduleHandler(service services.DrugScheduleService) *DrugScheduleHandler {
	return &DrugScheduleHandler{service: service}
}

// toDrugScheduleResponse adalah helper function untuk mengubah model ke DTO response.
func toDrugScheduleResponse(schedule models.DrugSchedule) dto.DrugScheduleResponseDTO {
	return dto.DrugScheduleResponseDTO{
		ID:           schedule.ID,
		UserID:       schedule.UserID,
		DrugName:     schedule.DrugName,
		Dose:         schedule.Dose,
		ScheduleDate: schedule.ScheduleDate.Format("2006-01-02"),
		At06:         schedule.At06,
		At12:         schedule.At12,
		At18:         schedule.At18,
		IsActive:     schedule.IsActive,
	}
}

// Create menangani pembuatan jadwal minum obat baru.
func (h *DrugScheduleHandler) Create(c *gin.Context) {
	var input dto.CreateDrugScheduleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Create(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create drug schedule", err.Error()))
		return
	}

	response := utils.SuccessResponse("Drug schedule created successfully", toDrugScheduleResponse(schedule))
	c.JSON(http.StatusCreated, response)
}

// GetAll menangani pengambilan semua jadwal minum obat milik user.
func (h *DrugScheduleHandler) GetAll(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	schedules, err := h.service.FindAllByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch drug schedules", err.Error()))
		return
	}

	var responseDTOs []dto.DrugScheduleResponseDTO
	for _, s := range schedules {
		responseDTOs = append(responseDTOs, toDrugScheduleResponse(s))
	}

	response := utils.SuccessResponse("Drug schedules fetched successfully", responseDTOs)
	c.JSON(http.StatusOK, response)
}

// GetByID menangani pengambilan satu jadwal berdasarkan ID.
func (h *DrugScheduleHandler) GetByID(c *gin.Context) {
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

	response := utils.SuccessResponse("Drug schedule fetched successfully", toDrugScheduleResponse(schedule))
	c.JSON(http.StatusOK, response)
}

// Update menangani pembaruan data jadwal.
func (h *DrugScheduleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	var input dto.UpdateDrugScheduleDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Update(uint(id), uint(userID), input) // Kirim userID ke service
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update drug schedule", err.Error()))
		return
	}

	response := utils.SuccessResponse("Drug schedule updated successfully", toDrugScheduleResponse(schedule))
	c.JSON(http.StatusOK, response)
}

// Delete menangani penghapusan jadwal.
func (h *DrugScheduleHandler) Delete(c *gin.Context) {
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

	response := utils.SuccessResponse("Drug schedule deleted successfully", nil)
	c.JSON(http.StatusOK, response)
}
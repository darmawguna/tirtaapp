package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto" // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type HemodialysisMonitoringHandler struct {
	service services.HemodialysisMonitoringService
}

func NewHemodialysisMonitoringHandler(service services.HemodialysisMonitoringService) *HemodialysisMonitoringHandler {
	return &HemodialysisMonitoringHandler{service: service}
}

func toHemodialysisMonitoringResponse(m models.HemodialysisMonitoring) dto.HemodialysisMonitoringResponseDTO {
	return dto.HemodialysisMonitoringResponseDTO{
		ID:                     m.ID,
		UserID:                 m.UserID,
		HemodialysisScheduleID: m.HemodialysisScheduleID,
		// Access schedule date via the preloaded HemodialysisSchedule relation
		ScheduleDate: m.HemodialysisSchedule.ScheduleDate.Format("2006-01-02"),
		BPBefore:     m.BPBefore,
		BPAfter:      m.BPAfter,
		WeightBefore: m.WeightBefore,
		WeightAfter:  m.WeightAfter,
	}
}

func (h *HemodialysisMonitoringHandler) Create(c *gin.Context) {
	scheduleID, err := strconv.ParseUint(c.Param("schedule_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid schedule ID format", err.Error()))
		return
	}

	var input dto.CreateHemodialysisMonitoringDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	monitoring, err := h.service.CreateMonitoring(uint(userID), uint(scheduleID), input)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		// Check for unique constraint violation (already exists)
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.JSON(http.StatusConflict, utils.ErrorResponse("Monitoring data for this schedule already exists", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to save monitoring data", err.Error()))
		return
	}

	// We need to fetch the data again to preload the schedule date for the response
	savedMonitoring, _ := h.service.GetMonitoringByScheduleID(uint(userID), monitoring.HemodialysisScheduleID)

	c.JSON(http.StatusCreated, utils.SuccessResponse("Monitoring data saved", toHemodialysisMonitoringResponse(savedMonitoring)))
}

func (h *HemodialysisMonitoringHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	history, err := h.service.GetMonitoringHistory(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch monitoring history", err.Error()))
		return
	}

	var responseDTOs []dto.HemodialysisMonitoringResponseDTO
	for _, m := range history {
		responseDTOs = append(responseDTOs, toHemodialysisMonitoringResponse(m))
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Monitoring history fetched", responseDTOs))
}

// Add GetByScheduleID handler if needed
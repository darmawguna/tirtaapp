package handlers

import (
	"net/http"

	"github.com/darmawguna/tirtaapp.git/dto" // Sesuaikan path
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type FluidBalanceHandler struct {
	service services.FluidBalanceService
}

func NewFluidBalanceHandler(service services.FluidBalanceService) *FluidBalanceHandler {
	return &FluidBalanceHandler{service: service}
}

func toFluidBalanceResponse(log models.FluidBalanceLog) dto.FluidBalanceLogResponseDTO {
	return dto.FluidBalanceLogResponseDTO{
		ID:             log.ID,
		UserID:         log.UserID,
		LogDate:        log.LogDate.Format("2006-01-02"),
		IntakeCC:       log.IntakeCC,
		OutputCC:       log.OutputCC,
		BalanceCC:      log.BalanceCC,
		WarningMessage: log.WarningMessage,
	}
}

func (h *FluidBalanceHandler) CreateOrUpdate(c *gin.Context) {
	var input dto.CreateOrUpdateFluidLogDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	logEntry, err := h.service.CreateOrUpdateLog(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to save fluid log", err.Error()))
		return
	}

	response := utils.SuccessResponse("Fluid log saved successfully", toFluidBalanceResponse(logEntry))
	c.JSON(http.StatusOK, response) // Gunakan 200 OK karena ini bisa create atau update
}

func (h *FluidBalanceHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	logs, err := h.service.GetUserHistory(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch history", err.Error()))
		return
	}

	var responseDTOs []dto.FluidBalanceLogResponseDTO
	for _, log := range logs {
		responseDTOs = append(responseDTOs, toFluidBalanceResponse(log))
	}

	response := utils.SuccessResponse("History fetched successfully", responseDTOs)
	c.JSON(http.StatusOK, response)
}
package handlers

import (
	"net/http"
	"strconv"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type ComplaintHandler struct {
	complaintService services.ComplaintService
}

func NewComplaintHandler(complaintService services.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{complaintService: complaintService}
}

// POST /api/v1/complaints
func (h *ComplaintHandler) Create(c *gin.Context) {
	var input dto.CreateComplaintDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Validation failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	userID := c.MustGet("userID").(float64)

	resp, err := h.complaintService.ProcessComplaint(uint(userID), input)
	if err != nil {
		// validation/business errors -> 400
		response := utils.ErrorResponse("Failed to process complaint", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.SuccessResponse("Complaint processed successfully", resp)
	c.JSON(http.StatusCreated, response)
}

// GET /api/v1/complaints?phase=pre_hd|post_hd
func (h *ComplaintHandler) GetMyComplaints(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	var phasePtr *dto.ComplaintPhase
	if phase := c.Query("phase"); phase != "" {
		p := dto.ComplaintPhase(phase)
		phasePtr = &p
	}

	logs, err := h.complaintService.GetMyComplaints(uint(userID), phasePtr)
	if err != nil {
		response := utils.ErrorResponse("Failed to fetch complaint history", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.SuccessResponse("Complaint history fetched successfully", logs)
	c.JSON(http.StatusOK, response)
}

// GET /api/v1/complaints/:id
func (h *ComplaintHandler) GetDetailComplaint(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := utils.ErrorResponse("Invalid ID format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	complaint, err := h.complaintService.GetComplaintByID(uint(userID), uint(id64))
	if err != nil {
		// For security, treat not found/ownership mismatch as 404 in many designs.
		response := utils.ErrorResponse("Complaint not found", err.Error())
		c.JSON(http.StatusNotFound, response)
		return
	}

	response := utils.SuccessResponse("Complaint fetched successfully", complaint)
	c.JSON(http.StatusOK, response)
}

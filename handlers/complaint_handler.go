package handlers

import (
	"net/http"
	"strconv"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

// **ComplaintHandler** adalah struct yang menampung service untuk keluhan.
type ComplaintHandler struct {
	complaintService services.ComplaintService
}

// **NewComplaintHandler** adalah constructor untuk ComplaintHandler.
func NewComplaintHandler(complaintService services.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{complaintService: complaintService}
}

// **Create** menangani pembuatan log keluhan baru.
// Endpoint: POST /api/v1/complaints
func (h *ComplaintHandler) Create(c *gin.Context) {
	var input dto.CreateComplaintDTO

	// Binding dan validasi request body.
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Validation failed. Make sure 'complaints' is a non-empty array.", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Ambil ID user dari context yang di-set oleh middleware.
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk memproses keluhan dan mendapatkan pesan balasan.
	generatedMessage, err := h.complaintService.ProcessComplaint(uint(userID), input)
	if err != nil {
		response := utils.ErrorResponse("Failed to process complaint", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Kirim response sukses yang berisi pesan yang dihasilkan.
	response := utils.SuccessResponse("Complaint processed successfully", gin.H{
		"generated_message": generatedMessage,
	})
	c.JSON(http.StatusCreated, response)
}

// **GetMyComplaints** menangani permintaan untuk melihat riwayat keluhan user.
// Endpoint: GET /api/v1/complaints
func (h *ComplaintHandler) GetMyComplaints(c *gin.Context) {
	// Ambil ID user dari context.
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk mendapatkan data riwayat.
	logs, err := h.complaintService.GetMyComplaints(uint(userID))
	if err != nil {
		response := utils.ErrorResponse("Failed to fetch complaint history", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Kirim response sukses dengan data riwayat.
	response := utils.SuccessResponse("Complaint history fetched successfully", logs)
	c.JSON(http.StatusOK, response)
}

func (h *ComplaintHandler) GetDetailComplaint(c *gin.Context) {
	// Ambil ID user dari context.
	complaint_id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := utils.ErrorResponse("Invalid ID format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	complaint, err := h.complaintService.GetComplainById(uint(complaint_id))
	if err != nil {
		// Jika record tidak ditemukan, GORM akan memberikan error.
		response := utils.ErrorResponse("Fetching Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Kirim response sukses dengan data riwayat.
	response := utils.SuccessResponse("Complaint  fetched successfully", complaint)
	c.JSON(http.StatusOK, response)
}
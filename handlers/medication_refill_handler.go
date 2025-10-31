package handlers

import (
	"net/http"
	"strconv"
	"strings" // Pastikan import strings

	"github.com/darmawguna/tirtaapp.git/dto"          // Sesuaikan path
	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type MedicationRefillHandler struct {
	service services.MedicationRefillService
}

func NewMedicationRefillHandler(service services.MedicationRefillService) *MedicationRefillHandler {
	return &MedicationRefillHandler{service: service}
}

// Helper untuk konversi ke DTO response
func toMedicationRefillResponseDTO(schedule models.MedicationRefillSchedule) dto.MedicationRefillResponseDTO {
	return dto.MedicationRefillResponseDTO{
		ID:           schedule.ID,
		UserID:       schedule.UserID,
		RefillDate:   schedule.RefillDate.Format("2006-01-02"),
		IsActive:     schedule.IsActive,
	}
}

// --- Handler Functions ---

// Create: Menangani POST /api/v1/medication-refills
func (h *MedicationRefillHandler) Create(c *gin.Context) {
	var input dto.CreateMedicationRefillDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Create(uint(userID), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal membuat jadwal obat habis", err.Error()))
		return
	}

	response := utils.SuccessResponse("Jadwal obat habis berhasil dibuat", toMedicationRefillResponseDTO(schedule))
	c.JSON(http.StatusCreated, response)
}

// GetAll: Menangani GET /api/v1/medication-refills
func (h *MedicationRefillHandler) GetAll(c *gin.Context) {
	userID := c.MustGet("userID").(float64)
	schedules, err := h.service.FindAllByUserID(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal mengambil jadwal obat habis", err.Error()))
		return
	}

	var responseDTOs []dto.MedicationRefillResponseDTO
	for _, s := range schedules {
		responseDTOs = append(responseDTOs, toMedicationRefillResponseDTO(s))
	}

	response := utils.SuccessResponse("Jadwal obat habis berhasil diambil", responseDTOs)
	c.JSON(http.StatusOK, response)
}

// GetByID: Menangani GET /api/v1/medication-refills/:id
func (h *MedicationRefillHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Format ID tidak valid", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.FindByID(uint(id))
	if err != nil {
		// Asumsi service mengembalikan error spesifik
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal mengambil jadwal", err.Error()))
		return
	}

	// Verifikasi kepemilikan (meskipun service sudah, ini lapisan tambahan)
	if schedule.UserID != uint(userID) {
		c.JSON(http.StatusForbidden, utils.ErrorResponse("Anda tidak berwenang melihat jadwal ini", nil))
		return
	}

	response := utils.SuccessResponse("Jadwal obat habis berhasil diambil", toMedicationRefillResponseDTO(schedule))
	c.JSON(http.StatusOK, response)
}

// Update: Menangani PUT /api/v1/medication-refills/:id
func (h *MedicationRefillHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Format ID tidak valid", err.Error()))
		return
	}

	var input dto.UpdateMedicationRefillDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	schedule, err := h.service.Update(uint(id), uint(userID), input)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal memperbarui jadwal", err.Error()))
		return
	}

	response := utils.SuccessResponse("Jadwal obat habis berhasil diperbarui", toMedicationRefillResponseDTO(schedule))
	c.JSON(http.StatusOK, response)
}

// Delete: Menangani DELETE /api/v1/medication-refills/:id
func (h *MedicationRefillHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Format ID tidak valid", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	err = h.service.Delete(uint(id), uint(userID))
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menghapus jadwal", err.Error()))
		return
	}

	response := utils.SuccessResponse("Jadwal obat habis berhasil dinonaktifkan", nil) // Pesan diubah menjadi 'dinonaktifkan'
	c.JSON(http.StatusOK, response)
}
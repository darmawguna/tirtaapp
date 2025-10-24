package handlers

import (
	"net/http"
	"strconv"
	"strings" // Import strings untuk cek error

	"github.com/darmawguna/tirtaapp.git/dto"          // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

// HemodialysisMonitoringHandler mengelola request HTTP untuk pemantauan HD.
type HemodialysisMonitoringHandler struct {
	service services.HemodialysisMonitoringService
}

// NewHemodialysisMonitoringHandler adalah constructor.
func NewHemodialysisMonitoringHandler(service services.HemodialysisMonitoringService) *HemodialysisMonitoringHandler {
	return &HemodialysisMonitoringHandler{service: service}
}

// toHemodialysisMonitoringResponse mengonversi model ke DTO response.
func toHemodialysisMonitoringResponse(m models.HemodialysisMonitoring) dto.HemodialysisMonitoringResponseDTO {
	return dto.HemodialysisMonitoringResponseDTO{
		ID:             m.ID,
		UserID:         m.UserID,
		MonitoringDate: m.MonitoringDate.Format("2006-01-02"), // Format tanggal YYYY-MM-DD
		BPBefore:       m.BPBefore,
		BPAfter:        m.BPAfter,
		WeightBefore:   m.WeightBefore,
		WeightAfter:    m.WeightAfter,
	}
}

// Create menangani request POST /api/v1/hemodialysis-monitoring
func (h *HemodialysisMonitoringHandler) Create(c *gin.Context) {
	var input dto.CreateHemodialysisMonitoringDTO
	// Bind data JSON dari body request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	// Ambil userID dari context (dari AuthMiddleware)
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk membuat atau memperbarui data hari ini
	monitoring, err := h.service.CreateOrUpdateMonitoringForToday(uint(userID), input)
	if err != nil {
		// Tangani error spesifik dari service jika perlu (misal: duplikasi ditangani di repo/service)
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") { // Contoh error Postgres
			c.JSON(http.StatusConflict, utils.ErrorResponse("Data pemantauan untuk hari ini sudah ada", err.Error()))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal menyimpan data pemantauan", err.Error()))
		return
	}

	// Kirim response sukses dengan data yang disimpan/diperbarui
	c.JSON(http.StatusOK, utils.SuccessResponse("Data pemantauan berhasil disimpan", toHemodialysisMonitoringResponse(monitoring)))
}

// GetHistory menangani request GET /api/v1/hemodialysis-monitoring/history
func (h *HemodialysisMonitoringHandler) GetHistory(c *gin.Context) {
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk mengambil riwayat
	history, err := h.service.GetMonitoringHistory(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal mengambil riwayat pemantauan", err.Error()))
		return
	}

	// Konversi setiap item riwayat ke DTO response
	var responseDTOs []dto.HemodialysisMonitoringResponseDTO
	for _, m := range history {
		responseDTOs = append(responseDTOs, toHemodialysisMonitoringResponse(m))
	}

	// Kirim response sukses dengan daftar riwayat
	c.JSON(http.StatusOK, utils.SuccessResponse("Riwayat pemantauan berhasil diambil", responseDTOs))
}

func (h *HemodialysisMonitoringHandler) GetByID(c *gin.Context) {
	// Ambil ID monitoring dari parameter URL
	monitoringID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid monitoring ID format", err.Error()))
		return
	}

	// Ambil userID dari context
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk mendapatkan data dan verifikasi kepemilikan
	monitoring, err := h.service.GetMonitoringByID(uint(userID), uint(monitoringID))
	if err != nil {
		if strings.Contains(err.Error(), "tidak ditemukan") {
			c.JSON(http.StatusNotFound, utils.ErrorResponse(err.Error(), nil))
			return
		}
		if strings.Contains(err.Error(), "tidak berwenang") {
			c.JSON(http.StatusForbidden, utils.ErrorResponse(err.Error(), nil))
			return
		}
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Gagal mengambil data pemantauan", err.Error()))
		return
	}

	// Kirim response sukses
	c.JSON(http.StatusOK, utils.SuccessResponse("Data pemantauan berhasil diambil", toHemodialysisMonitoringResponse(monitoring)))
}
package handlers

import (
	"net/http"
	"strconv"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

// **QuizHandler** adalah struct yang menampung service untuk kuis.
type QuizHandler struct {
	quizService services.QuizService
}

// **NewQuizHandler** adalah constructor untuk QuizHandler.
func NewQuizHandler(quizService services.QuizService) *QuizHandler {
	return &QuizHandler{quizService: quizService}
}

// **Create** menangani pembuatan kuis baru.
// Endpoint: POST /api/v1/quizzes
func (h *QuizHandler) Create(c *gin.Context) {
	var input dto.CreateQuizDTO

	// Binding dan validasi request body.
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Create Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Ambil ID user (admin) dari context yang di-set oleh middleware.
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk membuat kuis.
	quiz, err := h.quizService.Create(input, uint(userID))
	if err != nil {
		response := utils.ErrorResponse("Create Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	c.JSON(http.StatusCreated, quiz)
}

// **GetAll** menangani pengambilan semua data kuis.
// Endpoint: GET /api/v1/quizzes
func (h *QuizHandler) GetAll(c *gin.Context) {
	quizzes, err := h.quizService.FindAll()
	if err != nil {
		response := utils.ErrorResponse("Fetching Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.SuccessResponse("Fetching Quiz successfully", quizzes)
	c.JSON(http.StatusCreated, response)
}

// **GetByID** menangani pengambilan satu kuis berdasarkan ID.
// Endpoint: GET /api/v1/quizzes/:id
func (h *QuizHandler) GetByID(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := utils.ErrorResponse("Fetching Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	quiz, err := h.quizService.FindByID(uint(id))
	if err != nil {
		// Jika record tidak ditemukan, GORM akan memberikan error.
		response := utils.ErrorResponse("Fetching Quiz failed", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}
	response := utils.SuccessResponse("Fetching Quiz successfully", quiz)
	c.JSON(http.StatusCreated, response)
}

// **Update** menangani pembaruan data kuis.
// Endpoint: PUT /api/v1/quizzes/:id
func (h *QuizHandler) Update(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := utils.ErrorResponse("Invalid ID format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var input dto.UpdateQuizDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		response := utils.ErrorResponse("Error Update", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	updatedQuiz, err := h.quizService.Update(uint(id), input)
	if err != nil {
		response := utils.ErrorResponse("Failed to update quiz", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.SuccessResponse("Fetching Quiz successfully", updatedQuiz)
	c.JSON(http.StatusCreated, response)
	
}

// **Delete** menangani penghapusan kuis.
// Endpoint: DELETE /api/v1/quizzes/:id
func (h *QuizHandler) Delete(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response := utils.ErrorResponse("Invalid ID format", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err := h.quizService.Delete(uint(id)); err != nil {
		response := utils.ErrorResponse("Failed to delete quiz", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.SuccessResponse("Quiz successfully deleted", struct{}{})
	c.JSON(http.StatusCreated, response)
}

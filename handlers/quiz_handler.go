package handlers

import (
	"net/http"
	"strconv"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ambil ID user (admin) dari context yang di-set oleh middleware.
	userID := c.MustGet("userID").(float64)

	// Panggil service untuk membuat kuis.
	quiz, err := h.quizService.Create(input, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quiz"})
		return
	}

	c.JSON(http.StatusCreated, quiz)
}

// **GetAll** menangani pengambilan semua data kuis.
// Endpoint: GET /api/v1/quizzes
func (h *QuizHandler) GetAll(c *gin.Context) {
	quizzes, err := h.quizService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch quizzes"})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}

// **GetByID** menangani pengambilan satu kuis berdasarkan ID.
// Endpoint: GET /api/v1/quizzes/:id
func (h *QuizHandler) GetByID(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	quiz, err := h.quizService.FindByID(uint(id))
	if err != nil {
		// Jika record tidak ditemukan, GORM akan memberikan error.
		c.JSON(http.StatusNotFound, gin.H{"error": "Quiz not found"})
		return
	}

	c.JSON(http.StatusOK, quiz)
}

// **Update** menangani pembaruan data kuis.
// Endpoint: PUT /api/v1/quizzes/:id
func (h *QuizHandler) Update(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input dto.UpdateQuizDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedQuiz, err := h.quizService.Update(uint(id), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quiz"})
		return
	}

	c.JSON(http.StatusOK, updatedQuiz)
}

// **Delete** menangani penghapusan kuis.
// Endpoint: DELETE /api/v1/quizzes/:id
func (h *QuizHandler) Delete(c *gin.Context) {
	// Ambil ID dari parameter URL.
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.quizService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete quiz"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quiz successfully deleted"})
}
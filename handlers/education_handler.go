package handlers

import (
	"net/http"
	"strconv"

	"github.com/darmawguna/tirtaapp.git/dto"
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
)

type EducationHandler struct {
	educationService services.EducationService
}

func NewEducationHandler(educationService services.EducationService) *EducationHandler {
	return &EducationHandler{educationService: educationService}
}

func (h *EducationHandler) Create(c *gin.Context) {
	var input dto.CreateEducationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)
	education, err := h.educationService.Create(input, uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create education", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse("Education created successfully", education))
}

func (h *EducationHandler) GetAll(c *gin.Context) {
	educations, err := h.educationService.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to fetch educations", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Educations fetched successfully", educations))
}

func (h *EducationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	education, err := h.educationService.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.ErrorResponse("Education not found", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Education fetched successfully", education))
}

func (h *EducationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	var input dto.UpdateEducationDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", err.Error()))
		return
	}

	updatedEducation, err := h.educationService.Update(uint(id), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update education", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Education updated successfully", updatedEducation))
}

func (h *EducationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	if err := h.educationService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete education", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.SuccessResponse("Education successfully deleted", nil))
}
package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"          // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model" // Adjust path
	"github.com/darmawguna/tirtaapp.git/services"
	"github.com/darmawguna/tirtaapp.git/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper" // Untuk base URL
)

const UPLOAD_PATH_EDUCATION = "./uploads/educations" // Definisikan path upload

type EducationHandler struct {
	educationService services.EducationService
}

func NewEducationHandler(educationService services.EducationService) *EducationHandler {
	// Pastikan direktori upload ada
	if err := os.MkdirAll(UPLOAD_PATH_EDUCATION, os.ModePerm); err != nil {
		log.Fatalf("FATAL: Tidak bisa membuat direktori upload edukasi: %v", err)
	}
	return &EducationHandler{educationService: educationService}
}

// Helper untuk konversi ke Response DTO
func toEducationResponseDTO(edu models.Education) dto.EducationResponseDTO {
	// Buat URL lengkap jika path relatif
	thumbnailUrl := edu.Thumbnail
	if !strings.HasPrefix(thumbnailUrl, "http") {
		// Asumsi Anda akan setup static file serving di /static
		// Ganti "http://localhost:8080" dengan base URL API Anda dari config
		baseUrl := viper.GetString("BASE_URL") // Tambahkan APP_BASE_URL=http://localhost:8080 ke .env
		if baseUrl == "" {
			baseUrl = "http://localhost:8080"
		} // Fallback
		thumbnailUrl = fmt.Sprintf("%s/static/educations/%s", baseUrl, filepath.Base(thumbnailUrl))
	}

	return dto.EducationResponseDTO{
		ID:        edu.ID,
		Name:      edu.Name,
		Url:       edu.Url,
		Thumbnail: thumbnailUrl, // Kirim URL lengkap ke frontend
		CreatedBy: edu.CreatedBy,
	}
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

// Create: Menangani upload file dan data form
func (h *EducationHandler) Create(c *gin.Context) {

	log.Println(">>> MASUK HANDLER: Create Education <<<")
	// Bind data form (Name, Url) dari multipart/form-data
	name := c.PostForm("name")
	url := c.PostForm("url")

	log.Printf(">>> Data Form: Name=%s, Url=%s", name, url)
	// Validasi manual sederhana
	if name == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", "Name is required"))
		return
	}
	if url == "" {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed", "Url is required"))
		return
	}

	input := dto.CreateEducationDTO{
		Name: name,
		Url:  url,
	}
	// Ambil file thumbnail dari form-data
	file, err := c.FormFile("thumbnail") // Nama field di form-data
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Thumbnail file is required", err.Error()))
		return
	}

	// Validasi sederhana (contoh: ukuran < 5MB)
	if file.Size > 5*1024*1024 { // 5 MB
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("File size exceeds 5MB limit", nil))
		return
	}
	// Anda bisa menambahkan validasi tipe file (misal: hanya gambar)

	// Simpan file dan dapatkan path relatifnya
	thumbnailPath, err := saveUploadedFile(c, file, UPLOAD_PATH_EDUCATION)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to save thumbnail", err.Error()))
		return
	}

	userID := c.MustGet("userID").(float64)

	// Panggil service dengan path thumbnail yang sudah disimpan
	education, err := h.educationService.Create(input, uint(userID), thumbnailPath)
	if err != nil {
		// Jika gagal simpan DB, coba hapus file yang sudah terupload
		go os.Remove(thumbnailPath)
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to create education", err.Error()))
		return
	}

	response := utils.SuccessResponse("Education created successfully", toEducationResponseDTO(education))
	c.JSON(http.StatusCreated, response)
}

// Update: Menangani update data dan optional upload file baru
func (h *EducationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	var input dto.UpdateEducationDTO
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Validation failed for Name/Url", err.Error()))
		return
	}

	var newThumbnailPath *string         // Pointer untuk menandakan ada file baru atau tidak
	file, err := c.FormFile("thumbnail") // Coba ambil file

	// Jika ada file baru di request
	if err == nil {
		// Validasi file baru
		if file.Size > 5*1024*1024 { /* ... handle error ... */
		}

		// Simpan file baru
		savedPath, saveErr := saveUploadedFile(c, file, UPLOAD_PATH_EDUCATION)
		if saveErr != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to save new thumbnail", saveErr.Error()))
			return
		}
		newThumbnailPath = &savedPath // Set pointer ke path baru
	} else if err != http.ErrMissingFile {
		// Error selain karena file tidak ada (misal: format request salah)
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Error processing thumbnail file", err.Error()))
		return
	}
	// Jika err == http.ErrMissingFile, berarti tidak ada thumbnail baru, newThumbnailPath tetap nil

	// Panggil service update (service akan menangani penghapusan file lama jika perlu)
	updatedEdu, err := h.educationService.Update(uint(id), input, newThumbnailPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to update education", err.Error()))
		return
	}

	response := utils.SuccessResponse("Education updated successfully", toEducationResponseDTO(updatedEdu))
	c.JSON(http.StatusOK, response)
}

// Delete: Handler memanggil service (service menghapus DB & file)
func (h *EducationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse("Invalid ID format", err.Error()))
		return
	}

	err = h.educationService.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse("Failed to delete education", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse("Education deleted successfully", nil))
}

// Implementasikan GetAll, GetByID dengan konversi ke DTO Response...

// --- Helper Function untuk Menyimpan File ---
// Mengembalikan path absolut file yang disimpan
func saveUploadedFile(c *gin.Context, file *multipart.FileHeader, destDirectory string) (string, error) {
	// Buat nama file unik (timestamp_namaasli.ext)
	ext := filepath.Ext(file.Filename)
	baseName := strings.TrimSuffix(file.Filename, ext)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), baseName, ext)

	// Pastikan nama file aman (opsional, tapi bagus)
	// filename = sanitize.BaseName(filename)

	dst := filepath.Join(destDirectory, filename)

	// Simpan file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", fmt.Errorf("gagal menyimpan file ke '%s': %w", dst, err)
	}
	log.Printf("File '%s' berhasil disimpan di: %s", file.Filename, dst)

	// Kembalikan path absolut untuk disimpan di DB dan digunakan os.Remove
	return dst, nil
}

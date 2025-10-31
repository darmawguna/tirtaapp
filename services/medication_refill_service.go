package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"          // Sesuaikan path
	models "github.com/darmawguna/tirtaapp.git/model" // Sesuaikan path
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/gorm"
)

// --- Interface ---
type MedicationRefillService interface {
	Create(userID uint, input dto.CreateMedicationRefillDTO) (models.MedicationRefillSchedule, error)
	FindAllByUserID(userID uint) ([]models.MedicationRefillSchedule, error)
	FindByID(id uint) (models.MedicationRefillSchedule, error)
	Update(id uint, userID uint, input dto.UpdateMedicationRefillDTO) (models.MedicationRefillSchedule, error)
	Delete(id uint, userID uint) error
}

// --- Struct ---
type medicationRefillService struct {
	repo         repositories.MedicationRefillRepository
	queueService QueueService // Untuk mengirim pesan notifikasi
}

// --- Constructor ---
func NewMedicationRefillService(repo repositories.MedicationRefillRepository, queueService QueueService) MedicationRefillService {
	return &medicationRefillService{repo: repo, queueService: queueService}
}

// --- Implementasi ---
func (s *medicationRefillService) Create(userID uint, input dto.CreateMedicationRefillDTO) (models.MedicationRefillSchedule, error) {
	refillDate, err := time.Parse("2006-01-02", input.RefillDate)
	if err != nil {
		return models.MedicationRefillSchedule{}, fmt.Errorf("format tanggal tidak valid: %w", err)
	}

	schedule := models.MedicationRefillSchedule{
		UserID:       userID,
		RefillDate:   refillDate,
		IsActive:     true, // Default aktif saat dibuat
	}

	createdSchedule, err := s.repo.Create(schedule)
	if err != nil {
		return models.MedicationRefillSchedule{}, fmt.Errorf("gagal menyimpan jadwal: %w", err)
	}

	// Kirim pesan ke RabbitMQ
	payload := ReminderMessage{
		ScheduleType: "OBAT_HABIS",
		ScheduleID:   createdSchedule.ID,
	}
	err = s.queueService.PublishMessage(payload)
	if err != nil {
		log.Printf("ERROR: Gagal publish pesan Obat Habis untuk ID %d: %v\n", createdSchedule.ID, err)
	}

	return createdSchedule, nil
}

func (s *medicationRefillService) FindAllByUserID(userID uint) ([]models.MedicationRefillSchedule, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *medicationRefillService) FindByID(id uint) (models.MedicationRefillSchedule, error) {
	return s.repo.FindByID(id)
}

// [PEMBARUAN] Logika Update
func (s *medicationRefillService) Update(id uint, userID uint, input dto.UpdateMedicationRefillDTO) (models.MedicationRefillSchedule, error) {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.MedicationRefillSchedule{}, errors.New("jadwal tidak ditemukan")
		}
		return models.MedicationRefillSchedule{}, err
	}

	if schedule.UserID != userID {
		return models.MedicationRefillSchedule{}, errors.New("unauthorized")
	}

	refillDate, err := time.Parse("2006-01-02", input.RefillDate)
	if err != nil {
		return models.MedicationRefillSchedule{}, fmt.Errorf("format tanggal tidak valid: %w", err)
	}

	// Cek apakah tanggal berubah
	dateChanged := !schedule.RefillDate.Equal(refillDate)

	// Terapkan pembaruan
	schedule.RefillDate = refillDate
	schedule.IsActive = *input.IsActive

	// Jika tanggal atau status aktif berubah, reset flag notifikasi
	// agar worker bisa mengirim notifikasi baru (jika masih aktif)
	if dateChanged || *input.IsActive {
		schedule.NotificationSent = false
	}

	updatedSchedule, err := s.repo.Update(schedule)
	if err != nil {
		return models.MedicationRefillSchedule{}, err
	}

	// Jika jadwal diperbarui dan masih aktif, kirim ulang pesan ke queue
	// Ini akan memastikan notifikasi dikirim pada tanggal baru.
	// Worker akan menangani pesan duplikat (lama & baru) dengan cek flag 'NotificationSent'.
	if updatedSchedule.IsActive {
		payload := ReminderMessage{
			ScheduleType: "OBAT_HABIS",
			ScheduleID:   updatedSchedule.ID,
		}
		err = s.queueService.PublishMessage(payload)
		if err != nil {
			log.Printf("ERROR: Gagal publish ulang pesan Obat Habis untuk ID %d: %v\n", updatedSchedule.ID, err)
		}
	}

	return updatedSchedule, nil
}

// [PEMBARUAN] Logika Delete (Soft Delete)
func (s *medicationRefillService) Delete(id uint, userID uint) error {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("jadwal tidak ditemukan")
		}
		return err
	}

	if schedule.UserID != userID {
		return errors.New("unauthorized")
	}

	// Daripada menghapus, kita set IsActive = false
	// Ini secara efektif "membatalkan" notifikasi di masa depan,
	// karena worker Anda (worker.go) sudah memeriksa flag ini.
	schedule.IsActive = false
	_, err = s.repo.Update(schedule) // Gunakan Update untuk soft delete
	if err != nil {
		return fmt.Errorf("gagal menonaktifkan jadwal: %w", err)
	}

	// Jika Anda benar-benar ingin menghapus dari DB:
	// return s.repo.Delete(id)
	// Logika worker Anda yang memeriksa gorm.ErrRecordNotFound sudah
	// cukup untuk menangani "pembatalan" notifikasi.

	return nil
}
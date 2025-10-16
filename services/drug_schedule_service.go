package services

import (
	"errors"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type DrugScheduleService interface {
	Create(userID uint, input dto.CreateDrugScheduleDTO) (models.DrugSchedule, error)
	FindAllByUserID(userID uint) ([]models.DrugSchedule, error)
	FindByID(id uint) (models.DrugSchedule, error)
	Update(id uint, userID uint, input dto.UpdateDrugScheduleDTO) (models.DrugSchedule, error)
	Delete(id uint, userID uint) error
}

type drugScheduleService struct {
	repo         repositories.DrugScheduleRepository
	queueService QueueService // Suntikkan QueueService
}

func NewDrugScheduleService(repo repositories.DrugScheduleRepository, queueService QueueService) DrugScheduleService {
	return &drugScheduleService{repo: repo, queueService: queueService}
}

func (s *drugScheduleService) Create(userID uint, input dto.CreateDrugScheduleDTO) (models.DrugSchedule, error) {
	scheduleDate, err := time.Parse("2006-01-02", input.ScheduleDate)
	if err != nil { return models.DrugSchedule{}, err }

	schedule := models.DrugSchedule{
		UserID:       userID,
		DrugName:     input.DrugName,
		Dose:         input.Dose,
		ScheduleDate: scheduleDate,
		At06:         input.At06,
		At12:         input.At12,
		At18:         input.At18,
	}

	createdSchedule, err := s.repo.Create(schedule)
	if err != nil { return models.DrugSchedule{}, err }

	// Kirim pesan ke Message Queue untuk setiap waktu yang aktif
	if createdSchedule.At06 {
		s.publishReminder(createdSchedule.ID, 6)
	}
	if createdSchedule.At12 {
		s.publishReminder(createdSchedule.ID, 12)
	}
	if createdSchedule.At18 {
		s.publishReminder(createdSchedule.ID, 18)
	}

	return createdSchedule, nil
}

// Fungsi publishReminder sekarang lebih sederhana
func (s *drugScheduleService) publishReminder(scheduleID uint, hour int) {
	payload := ReminderMessage{
		ScheduleType: "DRUG",
		ScheduleID:   scheduleID,
		TimeSlot:     hour,
	}

	// Kirim langsung tanpa delay
	err := s.queueService.PublishMessage(payload)
	if err != nil {
		log.Printf("ERROR: Failed to publish reminder message for schedule ID %d: %v\n", scheduleID, err)
	}
}
func (s *drugScheduleService) FindAllByUserID(userID uint) ([]models.DrugSchedule, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *drugScheduleService) FindByID(id uint) (models.DrugSchedule, error) {
	return s.repo.FindByID(id)
}

func (s *drugScheduleService) Update(id uint, userID uint, input dto.UpdateDrugScheduleDTO) (models.DrugSchedule, error) {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return models.DrugSchedule{}, err
	}

	// Otorisasi: Pastikan user hanya mengubah data miliknya
	if schedule.UserID != userID {
		return models.DrugSchedule{}, errors.New("unauthorized to update this schedule")
	}

	scheduleDate, err := time.Parse("2006-01-02", input.ScheduleDate)
	if err != nil {
		return models.DrugSchedule{}, err
	}

	schedule.DrugName = input.DrugName
	schedule.Dose = input.Dose
	schedule.ScheduleDate = scheduleDate
	schedule.At06 = input.At06
	schedule.At12 = input.At12
	schedule.At18 = input.At18
	schedule.IsActive = *input.IsActive

	// TODO: Tambahkan logika untuk membatalkan notifikasi lama dan membuat notifikasi baru.
	// Untuk saat ini, kita hanya memperbarui datanya.

	return s.repo.Update(schedule)
}

func (s *drugScheduleService) Delete(id uint, userID uint) error {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Otorisasi: Pastikan user hanya menghapus data miliknya
	if schedule.UserID != userID {
		return errors.New("unauthorized to delete this schedule")
	}
    
	// TODO: Tambahkan logika untuk membatalkan notifikasi yang sudah dijadwalkan di message queue.
	// Salah satu cara sederhana adalah membuat worker memeriksa status 'IsActive' sebelum mengirim notifikasi.

	return s.repo.Delete(id)
}







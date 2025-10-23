package services

import (
	"errors"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type ControlScheduleService interface {
	Create(userID uint, input dto.CreateControlScheduleDTO) (models.ControlSchedule, error)
	FindAllByUserID(userID uint) ([]models.ControlSchedule, error)
	FindByID(id uint) (models.ControlSchedule, error)
	Update(id uint, userID uint, input dto.UpdateControlScheduleDTO) (models.ControlSchedule, error)
	Delete(id uint, userID uint) error
}

type controlScheduleService struct {
	repo         repositories.ControlScheduleRepository
	queueService QueueService // Akan kita gunakan di langkah berikutnya
}

func NewControlScheduleService(repo repositories.ControlScheduleRepository, queueService QueueService) ControlScheduleService {
	return &controlScheduleService{repo: repo, queueService: queueService}
}

func (s *controlScheduleService) Create(userID uint, input dto.CreateControlScheduleDTO) (models.ControlSchedule, error) {
	controlDate, err := time.Parse("2006-01-02", input.ControlDate)
	if err != nil {
		return models.ControlSchedule{}, err
	}

	schedule := models.ControlSchedule{
		UserID:      userID,
		ControlDate: controlDate,
	}

	createdSchedule, err := s.repo.Create(schedule)
	if err != nil {
		return models.ControlSchedule{}, err
	}

	// [LOGIKA BARU] Kirim pesan ke RabbitMQ setelah berhasil membuat jadwal
	payload := ReminderMessage{
		ScheduleType: "KONTROL",
		ScheduleID:   createdSchedule.ID,
	}
	err = s.queueService.PublishMessage(payload)
	if err != nil {
		// Log error jika gagal mengirim pesan, tapi jangan sampai menggagalkan proses utama
		log.Printf("ERROR: Failed to publish reminder for control schedule ID %d: %v\n", createdSchedule.ID, err)
	}

	return createdSchedule, nil
}

func (s *controlScheduleService) FindAllByUserID(userID uint) ([]models.ControlSchedule, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *controlScheduleService) FindByID(id uint) (models.ControlSchedule, error) {
	return s.repo.FindByID(id)
}

func (s *controlScheduleService) Update(id uint, userID uint, input dto.UpdateControlScheduleDTO) (models.ControlSchedule, error) {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return models.ControlSchedule{}, err
	}
	if schedule.UserID != userID {
		return models.ControlSchedule{}, errors.New("unauthorized")
	}

	controlDate, err := time.Parse("2006-01-02", input.ControlDate)
	if err != nil {
		return models.ControlSchedule{}, err
	}

	schedule.ControlDate = controlDate
	schedule.IsActive = *input.IsActive

	return s.repo.Update(schedule)
}

func (s *controlScheduleService) Delete(id uint, userID uint) error {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if schedule.UserID != userID {
		return errors.New("unauthorized")
	}
	return s.repo.Delete(id)
}
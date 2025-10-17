package services

import (
	"errors"
	"log"
	"time"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type HemodialysisScheduleService interface {
	Create(userID uint, input dto.CreateHemodialysisScheduleDTO) (models.HemodialysisSchedule, error)
	FindAllByUserID(userID uint) ([]models.HemodialysisSchedule, error)
	FindByID(id uint) (models.HemodialysisSchedule, error)
	Update(id uint, userID uint, input dto.UpdateHemodialysisScheduleDTO) (models.HemodialysisSchedule, error)
	Delete(id uint, userID uint) error
}

type hemodialysisScheduleService struct {
	repo         repositories.HemodialysisScheduleRepository
	queueService QueueService
}

func NewHemodialysisScheduleService(repo repositories.HemodialysisScheduleRepository, queueService QueueService) HemodialysisScheduleService {
	return &hemodialysisScheduleService{repo: repo, queueService: queueService}
}

func (s *hemodialysisScheduleService) Create(userID uint, input dto.CreateHemodialysisScheduleDTO) (models.HemodialysisSchedule, error) {
	scheduleDate, err := time.Parse("2006-01-02", input.ScheduleDate)
	if err != nil {
		return models.HemodialysisSchedule{}, err
	}

	schedule := models.HemodialysisSchedule{
		UserID:       userID,
		ScheduleDate: scheduleDate,
		Notes:        input.Notes,
	}

	createdSchedule, err := s.repo.Create(schedule)
	if err != nil {
		return models.HemodialysisSchedule{}, err
	}

	// [LOGIKA BARU] Kirim pesan ke RabbitMQ
	payload := ReminderMessage{
		ScheduleType: "HEMODIALISA",
		ScheduleID:   createdSchedule.ID,
	}
	err = s.queueService.PublishMessage(payload)
	if err != nil {
		log.Printf("ERROR: Failed to publish reminder for hemodialysis schedule ID %d: %v\n", createdSchedule.ID, err)
	}

	return createdSchedule, nil
}

func (s *hemodialysisScheduleService) FindAllByUserID(userID uint) ([]models.HemodialysisSchedule, error) {
	return s.repo.FindAllByUserID(userID)
}

func (s *hemodialysisScheduleService) FindByID(id uint) (models.HemodialysisSchedule, error) {
	return s.repo.FindByID(id)
}

func (s *hemodialysisScheduleService) Update(id uint, userID uint, input dto.UpdateHemodialysisScheduleDTO) (models.HemodialysisSchedule, error) {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return models.HemodialysisSchedule{}, err
	}
	if schedule.UserID != userID {
		return models.HemodialysisSchedule{}, errors.New("unauthorized")
	}

	scheduleDate, err := time.Parse("2006-01-02", input.ScheduleDate)
	if err != nil {
		return models.HemodialysisSchedule{}, err
	}

	schedule.ScheduleDate = scheduleDate
	schedule.Notes = input.Notes
	schedule.IsActive = *input.IsActive

	return s.repo.Update(schedule)
}

func (s *hemodialysisScheduleService) Delete(id uint, userID uint) error {
	schedule, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if schedule.UserID != userID {
		return errors.New("unauthorized")
	}
	return s.repo.Delete(id)
}

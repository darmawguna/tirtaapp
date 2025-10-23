package services

import (
	"errors"
	"fmt"

	"github.com/darmawguna/tirtaapp.git/dto" // Adjust path
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
)

type HemodialysisMonitoringService interface {
	CreateMonitoring(userID uint, scheduleID uint, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error)
	GetMonitoringByScheduleID(userID, scheduleID uint) (models.HemodialysisMonitoring, error)
	GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error)
}

type hemodialysisMonitoringService struct {
	monitoringRepo repositories.HemodialysisMonitoringRepository
	scheduleRepo   repositories.HemodialysisScheduleRepository // Needed to verify schedule ownership
}

func NewHemodialysisMonitoringService(monitoringRepo repositories.HemodialysisMonitoringRepository, scheduleRepo repositories.HemodialysisScheduleRepository) HemodialysisMonitoringService {
	return &hemodialysisMonitoringService{monitoringRepo: monitoringRepo, scheduleRepo: scheduleRepo}
}

func (s *hemodialysisMonitoringService) CreateMonitoring(userID uint, scheduleID uint, input dto.CreateHemodialysisMonitoringDTO) (models.HemodialysisMonitoring, error) {
	// Verify the schedule exists and belongs to the user
	schedule, err := s.scheduleRepo.FindByID(scheduleID)
	if err != nil {
		return models.HemodialysisMonitoring{}, fmt.Errorf("schedule not found: %w", err)
	}
	if schedule.UserID != userID {
		return models.HemodialysisMonitoring{}, errors.New("unauthorized to add monitoring for this schedule")
	}

	monitoring := models.HemodialysisMonitoring{
		UserID:                 userID,
		HemodialysisScheduleID: scheduleID,
		BPBefore:               input.BPBefore,
		BPAfter:                input.BPAfter,
		WeightBefore:           input.WeightBefore,
		WeightAfter:            input.WeightAfter,
	}
	return s.monitoringRepo.Create(monitoring)
}

func (s *hemodialysisMonitoringService) GetMonitoringByScheduleID(userID, scheduleID uint) (models.HemodialysisMonitoring, error) {
	monitoring, err := s.monitoringRepo.FindByScheduleID(scheduleID)
	if err != nil {
		return models.HemodialysisMonitoring{}, err
	}
	// Verify ownership
	if monitoring.UserID != userID {
		return models.HemodialysisMonitoring{}, errors.New("unauthorized")
	}
	return monitoring, nil
}

func (s *hemodialysisMonitoringService) GetMonitoringHistory(userID uint) ([]models.HemodialysisMonitoring, error) {
	// Fetch last 10 records for history, for example
	return s.monitoringRepo.FindHistoryByUserID(userID, 10)
}
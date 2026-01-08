package services

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/darmawguna/tirtaapp.git/dto"
	models "github.com/darmawguna/tirtaapp.git/model"
	"github.com/darmawguna/tirtaapp.git/repositories"
	"gorm.io/datatypes"
)

type ComplaintService interface {
	ProcessComplaint(userID uint, input dto.CreateComplaintDTO) (dto.ComplaintLogResponse, error)
	GetMyComplaints(userID uint, phase *dto.ComplaintPhase) ([]dto.ComplaintLogResponse, error)
	GetComplaintByID(userID uint, id uint) (dto.ComplaintLogResponse, error)
}

type complaintService struct {
	complaintRepo repositories.ComplaintRepository
}

func NewComplaintService(complaintRepo repositories.ComplaintRepository) ComplaintService {
	return &complaintService{complaintRepo: complaintRepo}
}

func (s *complaintService) ProcessComplaint(userID uint, input dto.CreateComplaintDTO) (dto.ComplaintLogResponse, error) {
	// phase validation
	if err := validatePhase(input.Phase); err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	// catalog map: code -> label
	codeToLabel, err := catalogMaps(input.Phase)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	// sanitize + validate complaints
	codes, err := sanitizeAndValidateComplaintCodes(input.Complaints, codeToLabel)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	// enforce "lainnya" rules
	_, hasOther := containsCode(codes, "keluhan_lainnya")
	if hasOther {
		if input.OtherText == nil || strings.TrimSpace(*input.OtherText) == "" {
			return dto.ComplaintLogResponse{}, fmt.Errorf("other_text is required when keluhan_lainnya selected")
		}
		trimmed := strings.TrimSpace(*input.OtherText)
		if len(trimmed) > 500 {
			return dto.ComplaintLogResponse{}, fmt.Errorf("other_text too long (max 500)")
		}
		input.OtherText = &trimmed
	} else {
		// optional strictness: reject other_text if not needed
		if input.OtherText != nil && strings.TrimSpace(*input.OtherText) != "" {
			return dto.ComplaintLogResponse{}, fmt.Errorf("other_text provided but keluhan_lainnya not selected")
		}
		input.OtherText = nil
	}

	// business message (fix rule: >1 => urgent)
	generatedMessage := generateAdviceMessage(len(codes))

	// store JSON array of codes
	complaintsJSON, err := json.Marshal(codes)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	log := models.HemodialysisComplaint{
		UserID:     userID,
		Phase:      string(input.Phase),
		Complaints: datatypes.JSON(complaintsJSON),
		OtherText:  input.OtherText,
		Message:    generatedMessage,
	}

	saved, err := s.complaintRepo.Create(log)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	return mapModelToResponse(saved, codeToLabel)
}

func (s *complaintService) GetMyComplaints(userID uint, phase *dto.ComplaintPhase) ([]dto.ComplaintLogResponse, error) {
	var logs []models.HemodialysisComplaint
	var err error

	if phase != nil {
		if e := validatePhase(*phase); e != nil {
			return nil, e
		}
		logs, err = s.complaintRepo.FindByUserIDAndPhase(userID, string(*phase))
	} else {
		logs, err = s.complaintRepo.FindByUserID(userID)
	}
	if err != nil {
		return nil, err
	}

	out := make([]dto.ComplaintLogResponse, 0, len(logs))
	for _, l := range logs {
		ph := dto.ComplaintPhase(l.Phase)
		codeToLabel, e := catalogMaps(ph)
		if e != nil {
			// if phase in DB unexpected, fail loud (data corruption)
			return nil, e
		}
		resp, e := mapModelToResponse(l, codeToLabel)
		if e != nil {
			return nil, e
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *complaintService) GetComplaintByID(userID uint, id uint) (dto.ComplaintLogResponse, error) {
	log, err := s.complaintRepo.FindByIDAndUserID(id, userID)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	ph := dto.ComplaintPhase(log.Phase)
	codeToLabel, err := catalogMaps(ph)
	if err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	return mapModelToResponse(log, codeToLabel)
}

// ----- helpers -----

func generateAdviceMessage(count int) string {
	if count <= 1 {
		return "Konsultasikan keluhan bapak/ibu kepada dokter/perawat yang bertugas atau hubungi petugas pada link TANYA PETUGAS"
	}
	return "Segera konsultasikan keluhan bapak/ibu ke poliklinik atau faskes terdekat / hubungi petugas"
}

func sanitizeAndValidateComplaintCodes(in []string, codeToLabel map[string]string) ([]string, error) {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))

	for _, raw := range in {
		code := normalizeCode(raw)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue // dedupe silently (recommended)
		}
		if _, ok := codeToLabel[code]; !ok {
			return nil, fmt.Errorf("invalid complaint code for this phase: %s", code)
		}
		seen[code] = struct{}{}
		out = append(out, code)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("complaints must contain at least 1 valid item")
	}
	return out, nil
}

func containsCode(list []string, code string) (int, bool) {
	for i, v := range list {
		if v == code {
			return i, true
		}
	}
	return -1, false
}

func mapModelToResponse(m models.HemodialysisComplaint, codeToLabel map[string]string) (dto.ComplaintLogResponse, error) {
	var codes []string
	if err := json.Unmarshal(m.Complaints, &codes); err != nil {
		return dto.ComplaintLogResponse{}, err
	}

	items := make([]dto.ComplaintItem, 0, len(codes))
	for _, c := range codes {
		label, ok := codeToLabel[c]
		if !ok {
			// DB contains code not in catalog -> data corruption/version mismatch
			label = c
		}
		items = append(items, dto.ComplaintItem{
			Code:  c,
			Label: label,
		})
	}

	return dto.ComplaintLogResponse{
		ID:               m.ID,
		Phase:            dto.ComplaintPhase(m.Phase),
		Complaints:       items,
		OtherText:        m.OtherText,
		GeneratedMessage: m.Message,
		CreatedAt:        m.CreatedAt,
	}, nil
}

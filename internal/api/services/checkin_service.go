package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinService struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewCheckinService(db *gorm.DB, cfg *config.Config) *CheckinService {
	return &CheckinService{db: db, cfg: cfg}
}

func (s *CheckinService) StartCheckin(patientID uuid.UUID, scheduleID *uuid.UUID) (*models.Checkin, error) {
	var patient models.User
	if err := s.db.First(&patient, "id = ? AND role = ?", patientID, enums.UserRolePatient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	var pat models.Patient
	if err := s.db.First(&pat, "user_id = ?", patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	if _, err := s.findActiveCheckin(pat.ID); err == nil {
		return nil, errs.ErrActiveCheckinExists
	} else if !errors.Is(err, errs.ErrNoActiveCheckin) {
		return nil, err
	}

	checkin := models.Checkin{
		PatientID:   pat.ID,
		ScheduleID:  scheduleID,
		Status:      enums.CheckinStatusInProgress,
		InitiatedAt: time.Now(),
	}

	if err := s.db.Create(&checkin).Error; err != nil {
		return nil, err
	}

	return &checkin, nil
}

func (s *CheckinService) EndCheckin(patientUserID uuid.UUID) (*models.Checkin, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", patientUserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	// Find the active check-in instead of any check-in
	checkin, err := s.findActiveCheckin(patient.ID)
	if err != nil {
		return nil, err
	}

	completedAt := time.Now()
	if err := s.db.Model(checkin).Updates(map[string]interface{}{
		"status":       enums.CheckinStatusCompleted,
		"completed_at": &completedAt,
	}).Error; err != nil {
		return nil, err
	}

	checkin.Status = enums.CheckinStatusCompleted
	checkin.CompletedAt = &completedAt

	return checkin, nil
}

func (s *CheckinService) GetActiveCheckin(patientID uuid.UUID) (*models.Checkin, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return s.findActiveCheckin(patient.ID)
}

func (s *CheckinService) GetByID(checkinID uuid.UUID) (*models.Checkin, error) {
	var checkin models.Checkin
	if err := s.db.First(&checkin, "id = ?", checkinID).Error; err != nil {
		return nil, err
	}
	return &checkin, nil
}

func (s *CheckinService) AddQuestions(checkinID uuid.UUID, questions []interface{}) (*models.Checkin, error) {
	return s.appendToArrayField(checkinID, "questions", questions)
}

func (s *CheckinService) AddAnswers(checkinID uuid.UUID, answers []interface{}) (*models.Checkin, error) {
	return s.appendToArrayField(checkinID, "answers", answers)
}

func (s *CheckinService) ListCompletedByPatient(patientID uuid.UUID) ([]models.Checkin, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	var checkins []models.Checkin
	if err := s.db.Where("patient_id = ? AND status = ?", patient.ID, enums.CheckinStatusCompleted).
		Order("initiated_at DESC").
		Find(&checkins).Error; err != nil {
		return nil, err
	}
	return checkins, nil
}

func (s *CheckinService) ReviewCheckin(checkinID, doctorID uuid.UUID, doctorNotes *string) (*models.Checkin, error) {
	if err := s.db.First(&models.User{}, "id = ? AND role = ?", doctorID, enums.UserRoleDoctor).Error; err != nil {
		return nil, err
	}

	var checkin models.Checkin
	if err := s.db.First(&checkin, "id = ?", checkinID).Error; err != nil {
		return nil, err
	}

	if checkin.AIAnalysis == nil {
		return nil, errs.ErrCheckinNotAnalyzed
	}

	reviewedAt := time.Now()
	updates := map[string]interface{}{
		"reviewed_by": doctorID,
		"reviewed_at": reviewedAt,
	}
	if doctorNotes != nil {
		updates["doctor_notes"] = doctorNotes
	}

	if err := s.db.Model(&checkin).Updates(updates).Error; err != nil {
		return nil, err
	}

	// acknowledge related alerts for this checkin
	ackUpdates := map[string]interface{}{
		"is_acknowledged": true,
		"acknowledged_by": doctorID,
		"acknowledged_at": reviewedAt,
	}
	if err := s.db.Model(&models.Alert{}).
		Where("checkin_id = ? AND is_acknowledged = ?", checkinID, false).
		Updates(ackUpdates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&checkin, "id = ?", checkin.ID).Error; err != nil {
		return nil, err
	}

	return &checkin, nil
}

type CheckinAIUpdate struct {
	AIAnalysis    *models.JSONB
	MedicalStatus *enums.MedicalStatus
	RiskScore     *int
	Alert         *CheckinAIAlertInput
}

type CheckinAIAlertInput struct {
	Severity  enums.AlertSeverity
	AlertType enums.AlertType
	Title     string
	Message   string
	Details   *models.JSONB
}

func (s *CheckinService) UpdateAIFields(checkinID uuid.UUID, input CheckinAIUpdate) (*models.Checkin, error) {
	var checkin models.Checkin
	if err := s.db.First(&checkin, "id = ?", checkinID).Error; err != nil {
		return nil, err
	}

	if checkin.Status != enums.CheckinStatusCompleted {
		return nil, errs.ErrCheckinNotCompleted
	}

	updates := map[string]interface{}{}
	if input.AIAnalysis != nil {
		updates["ai_analysis"] = *input.AIAnalysis
	}
	if input.MedicalStatus != nil {
		updates["medical_status"] = *input.MedicalStatus
	}
	if input.RiskScore != nil {
		updates["risk_score"] = input.RiskScore
	}

	if len(updates) == 0 {
		return &checkin, nil
	}

	if err := s.db.Model(&checkin).Updates(updates).Error; err != nil {
		return nil, err
	}

	// create alert if requested
	if input.Alert != nil {
		if err := s.createAlertFromAI(checkin, *input.Alert); err != nil {
			return nil, err
		}
	}

	if err := s.db.First(&checkin, "id = ?", checkin.ID).Error; err != nil {
		return nil, err
	}

	return &checkin, nil
}

func (s *CheckinService) createAlertFromAI(checkin models.Checkin, alertInput CheckinAIAlertInput) error {
	// minimal validation
	if alertInput.Severity == "" || alertInput.AlertType == "" || alertInput.Title == "" || alertInput.Message == "" {
		return errs.ErrMissingAlertFields
	}

	var patient models.Patient
	if err := s.db.First(&patient, "id = ?", checkin.PatientID).Error; err != nil {
		return err
	}

	alert := models.Alert{
		PatientID: patient.ID,
		CheckinID: &checkin.ID,
		Severity:  alertInput.Severity,
		AlertType: alertInput.AlertType,
		Title:     alertInput.Title,
		Message:   alertInput.Message,
	}
	if alertInput.Details != nil {
		alert.Details = *alertInput.Details
	}

	return s.db.Create(&alert).Error
}

func (s *CheckinService) findActiveCheckin(patientID uuid.UUID) (*models.Checkin, error) {
	var checkin models.Checkin
	if err := s.db.Where("patient_id = ? AND status IN ?", patientID, activeCheckinStatuses()).
		Order("initiated_at DESC").
		First(&checkin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNoActiveCheckin
		}
		return nil, err
	}
	return &checkin, nil
}

func (s *CheckinService) appendToArrayField(checkinID uuid.UUID, field string, items []interface{}) (*models.Checkin, error) {
	var checkin models.Checkin
	if err := s.db.First(&checkin, "id = ?", checkinID).Error; err != nil {
		return nil, err
	}

	if !isActiveCheckinStatus(checkin.Status) {
		return nil, errs.ErrCheckinNotActive
	}

	var current models.JSONB
	switch field {
	case "questions":
		current = checkin.Questions
	case "answers":
		current = checkin.Answers
	default:
		return nil, errors.New("unsupported field")
	}

	updated, err := appendJSONBArray(current, items)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(&checkin).Update(field, updated).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&checkin, "id = ?", checkinID).Error; err != nil {
		return nil, err
	}

	return &checkin, nil
}

func (s *CheckinService) StartManualCheckin(patientID uuid.UUID) (*models.Checkin, error) {
	var patient models.User
	if err := s.db.First(&patient, "id = ? AND role = ?", patientID, enums.UserRolePatient).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	var pat models.Patient
	if err := s.db.First(&pat, "user_id = ?", patientID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}

	var schedule models.CheckinSchedule
	if err := s.db.Where("patient_id = ? AND is_active = ?", pat.ID, true).First(&schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNoActiveSchedule
		}
		return nil, err
	}

	checkin, err := s.StartCheckin(patientID, &schedule.ID)
	if err != nil {
		return nil, err
	}

	// making request to external AI service
	url := fmt.Sprintf("%s/%s", s.cfg.Internal.TgBotURL, patient.ID)

	// create HTTP request
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// send request with timeout
	client := &http.Client{
		Timeout: 1 * time.Minute,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to bot service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("bot service returned error status: %d; body: %v", resp.StatusCode, string(body))
	}

	return checkin, nil
}

func appendJSONBArray(existing models.JSONB, additions []interface{}) (models.JSONB, error) {
	var arr []map[string]interface{}
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &arr); err != nil {
			return nil, err
		}
	}

	// Build a map of existing items by seq
	seqMap := make(map[float64]int) // seq -> index in arr
	for i, item := range arr {
		if seq, ok := item["seq"].(float64); ok {
			seqMap[seq] = i
		}
	}

	// Merge or append new items
	for _, add := range additions {
		addMap, ok := add.(map[string]interface{})
		if !ok {
			arr = append(arr, map[string]interface{}{"data": add})
			continue
		}

		if seq, exists := addMap["seq"].(float64); exists {
			if idx, found := seqMap[seq]; found {
				// Update existing item with same seq
				arr[idx] = addMap
			} else {
				// New seq, append
				arr = append(arr, addMap)
				seqMap[seq] = len(arr) - 1
			}
		} else {
			// No seq field, just append
			arr = append(arr, addMap)
		}
	}

	return models.NewJSONB(arr)
}

func activeCheckinStatuses() []enums.CheckinStatus {
	return []enums.CheckinStatus{
		enums.CheckinStatusPending,
		enums.CheckinStatusInProgress,
	}
}

func isActiveCheckinStatus(status enums.CheckinStatus) bool {
	for _, active := range activeCheckinStatuses() {
		if status == active {
			return true
		}
	}
	return false
}

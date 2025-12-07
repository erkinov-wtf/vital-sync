package services

import (
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertService struct {
	db *gorm.DB
}

func NewAlertService(db *gorm.DB) *AlertService {
	return &AlertService{db: db}
}

type CreateAlertInput struct {
	PatientID uuid.UUID
	CheckinID *uuid.UUID
	Severity  enums.AlertSeverity
	AlertType enums.AlertType
	Title     string
	Message   string
	Details   *models.JSONB
}

func (s *AlertService) Create(input CreateAlertInput) (*models.Alert, error) {
	// ensure patient exists (and implicitly doctor relationship)
	if err := s.db.First(&models.Patient{}, "id = ?", input.PatientID).Error; err != nil {
		return nil, err
	}

	alert := models.Alert{
		PatientID: input.PatientID,
		CheckinID: input.CheckinID,
		Severity:  input.Severity,
		AlertType: input.AlertType,
		Title:     input.Title,
		Message:   input.Message,
	}
	if input.Details != nil {
		alert.Details = *input.Details
	}

	if err := s.db.Create(&alert).Error; err != nil {
		return nil, err
	}

	return &alert, nil
}

func (s *AlertService) ListByDoctor(doctorID uuid.UUID, includeAcknowledged bool) ([]models.Alert, error) {
	// ensure doctor exists
	if err := s.db.First(&models.User{}, "id = ? AND role = ?", doctorID, enums.UserRoleDoctor).Error; err != nil {
		return nil, err
	}

	query := s.db.Model(&models.Alert{}).
		Joins("JOIN patients p ON p.id = alerts.patient_id").
		Where("p.doctor_id = ?", doctorID).
		Preload("Patient")
	if !includeAcknowledged {
		query = query.Where("alerts.is_acknowledged = ?", false)
	}

	var alerts []models.Alert
	if err := query.Order("alerts.created_at DESC").Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}

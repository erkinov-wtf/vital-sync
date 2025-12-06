package services

import (
	"errors"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/erkinov-wtf/vital-sync/internal/pkg/errs"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinScheduleService struct {
	db *gorm.DB
}

func NewCheckinScheduleService(db *gorm.DB) *CheckinScheduleService {
	return &CheckinScheduleService{db: db}
}

type CreateScheduleInput struct {
	PatientID     uuid.UUID
	Frequency     enums.ScheduleFrequency
	TimeSlots     []time.Time
	Timezone      *string
	IsActive      *bool
	NextCheckinAt *time.Time
}

func (s *CheckinScheduleService) Create(input CreateScheduleInput) (*models.CheckinSchedule, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", input.PatientID).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&models.Patient{}, "id = ?", patient.ID).Error; err != nil {
		return nil, err
	}

	// ensure unique per patient
	var existing models.CheckinSchedule
	if err := s.db.First(&existing, "patient_id = ?", patient.ID).Error; err == nil {
		return nil, errs.ErrScheduleExists
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	schedule := models.CheckinSchedule{
		PatientID: patient.ID,
		Frequency: input.Frequency,
		TimeSlots: models.TimeArray(input.TimeSlots),
	}

	if input.Timezone != nil {
		schedule.Timezone = *input.Timezone
	}
	if input.IsActive != nil {
		schedule.IsActive = *input.IsActive
	}
	if input.NextCheckinAt != nil {
		schedule.NextCheckinAt = input.NextCheckinAt
	}

	if err := s.db.Create(&schedule).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (s *CheckinScheduleService) GetByID(id uuid.UUID) (*models.CheckinSchedule, error) {
	var schedule models.CheckinSchedule
	if err := s.db.First(&schedule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (s *CheckinScheduleService) List(includeInactive bool, patientID *uuid.UUID) ([]models.CheckinSchedule, error) {
	query := s.db.Model(&models.CheckinSchedule{})
	if patientID != nil {
		query = query.Where("patient_id = ?", *patientID)
	}
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	var schedules []models.CheckinSchedule
	if err := query.Order("created_at DESC").Find(&schedules).Error; err != nil {
		return nil, err
	}
	return schedules, nil
}

type UpdateScheduleInput struct {
	Frequency     *enums.ScheduleFrequency
	TimeSlots     *[]time.Time
	Timezone      *string
	IsActive      *bool
	NextCheckinAt *time.Time
}

func (s *CheckinScheduleService) Update(id uuid.UUID, input UpdateScheduleInput) (*models.CheckinSchedule, error) {
	var schedule models.CheckinSchedule
	if err := s.db.First(&schedule, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Frequency != nil {
		updates["frequency"] = *input.Frequency
	}
	if input.TimeSlots != nil {
		updates["time_slots"] = models.TimeArray(*input.TimeSlots)
	}
	if input.Timezone != nil {
		updates["timezone"] = *input.Timezone
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.NextCheckinAt != nil {
		updates["next_checkin_at"] = input.NextCheckinAt
	}

	if len(updates) == 0 {
		return &schedule, nil
	}

	if err := s.db.Model(&schedule).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&schedule, "id = ?", schedule.ID).Error; err != nil {
		return nil, err
	}

	return &schedule, nil
}

func (s *CheckinScheduleService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.CheckinSchedule{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

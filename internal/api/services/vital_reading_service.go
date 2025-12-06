package services

import (
	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VitalReadingService struct {
	db *gorm.DB
}

func NewVitalReadingService(db *gorm.DB) *VitalReadingService {
	return &VitalReadingService{db: db}
}

type CreateVitalReadingInput struct {
	CheckinID uuid.UUID
	PatientID uuid.UUID
	VitalType enums.VitalType
	Unit      *enums.VitalUnit

	ValueNumeric *float64
	ValueText    *string

	IsAbnormal            *bool
	DeviationFromBaseline *float64
}

func (s *VitalReadingService) Create(input CreateVitalReadingInput) (*models.VitalReading, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", input.PatientID).Error; err != nil {
		return nil, err
	}

	if err := s.ensurePatientExists(patient.ID); err != nil {
		return nil, err
	}

	if err := s.ensureCheckinExists(input.CheckinID); err != nil {
		return nil, err
	}

	reading := models.VitalReading{
		CheckinID: input.CheckinID,
		PatientID: patient.ID,
		VitalType: input.VitalType,
		Unit:      input.Unit,
	}

	if input.ValueNumeric != nil {
		reading.ValueNumeric = input.ValueNumeric
	}
	if input.ValueText != nil {
		reading.ValueText = input.ValueText
	}
	if input.IsAbnormal != nil {
		reading.IsAbnormal = *input.IsAbnormal
	}
	if input.DeviationFromBaseline != nil {
		reading.DeviationFromBaseline = input.DeviationFromBaseline
	}

	if err := s.db.Create(&reading).Error; err != nil {
		return nil, err
	}

	return &reading, nil
}

func (s *VitalReadingService) GetByID(id uuid.UUID) (*models.VitalReading, error) {
	var reading models.VitalReading
	if err := s.db.First(&reading, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reading, nil
}

func (s *VitalReadingService) List(patientID, checkinID *uuid.UUID, vitalType *enums.VitalType, onlyAbnormal bool) ([]models.VitalReading, error) {
	query := s.db.Model(&models.VitalReading{})
	if patientID != nil {
		query = query.Where("patient_id = ?", *patientID)
	}
	if checkinID != nil {
		query = query.Where("checkin_id = ?", *checkinID)
	}
	if vitalType != nil {
		query = query.Where("vital_type = ?", *vitalType)
	}
	if onlyAbnormal {
		query = query.Where("is_abnormal = ?", true)
	}

	var readings []models.VitalReading
	if err := query.Order("created_at DESC").Find(&readings).Error; err != nil {
		return nil, err
	}

	return readings, nil
}

type UpdateVitalReadingInput struct {
	VitalType             *enums.VitalType
	Unit                  *enums.VitalUnit
	ValueNumeric          *float64
	ValueText             *string
	IsAbnormal            *bool
	DeviationFromBaseline *float64
}

func (s *VitalReadingService) Update(id uuid.UUID, input UpdateVitalReadingInput) (*models.VitalReading, error) {
	var reading models.VitalReading
	if err := s.db.First(&reading, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.VitalType != nil {
		updates["vital_type"] = *input.VitalType
	}
	if input.Unit != nil {
		updates["unit"] = *input.Unit
	}
	if input.ValueNumeric != nil {
		updates["value_numeric"] = input.ValueNumeric
	}
	if input.ValueText != nil {
		updates["value_text"] = input.ValueText
	}
	if input.IsAbnormal != nil {
		updates["is_abnormal"] = *input.IsAbnormal
	}
	if input.DeviationFromBaseline != nil {
		updates["deviation_from_baseline"] = input.DeviationFromBaseline
	}

	if len(updates) == 0 {
		return &reading, nil
	}

	if err := s.db.Model(&reading).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&reading, "id = ?", reading.ID).Error; err != nil {
		return nil, err
	}

	return &reading, nil
}

func (s *VitalReadingService) Delete(id uuid.UUID) error {
	result := s.db.Delete(&models.VitalReading{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *VitalReadingService) ensurePatientExists(id uuid.UUID) error {
	return s.db.First(&models.Patient{}, "id = ?", id).Error
}

func (s *VitalReadingService) ensureCheckinExists(id uuid.UUID) error {
	if err := s.db.First(&models.Checkin{}, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

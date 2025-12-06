package services

import (
	"errors"
	"log/slog"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db   *gorm.DB
	logg *slog.Logger
}

func NewUserService(db *gorm.DB, lgr *slog.Logger) *UserService {
	return &UserService{
		db:   db,
		logg: lgr,
	}
}

// Doctor flows

func (s *UserService) CreateDoctor(doctor *models.User, orgID uuid.UUID) (*models.User, *models.OrganizationDoctor, error) {
	doctor.Role = enums.UserRoleDoctor
	user, orgDoc, err := s.createAndAssignDoctor(doctor, orgID)
	if err != nil {
		s.logg.Error("couldn't create new doctor and assign", err.Error())
		return nil, nil, err
	}

	return user, orgDoc, nil
}

func (s *UserService) GetDoctorByID(id uuid.UUID) (*models.User, error) {
	var doctor models.User
	if err := s.db.Where("role = ?", enums.UserRoleDoctor).First(&doctor, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (s *UserService) ListDoctors(includeInactive bool) ([]models.User, error) {
	var doctors []models.User
	query := s.db.Where("role = ?", enums.UserRoleDoctor)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Order("created_at DESC").Find(&doctors).Error; err != nil {
		return nil, err
	}
	return doctors, nil
}

type DoctorUpdate struct {
	PhoneNumber      *string
	FirstName        *string
	LastName         *string
	Gender           *enums.Gender
	IsActive         *bool
	Password         *string
	TelegramUsername *string
}

func (s *UserService) UpdateDoctor(id uuid.UUID, changes DoctorUpdate) (*models.User, error) {
	var doctor models.User
	if err := s.db.Where("role = ?", enums.UserRoleDoctor).First(&doctor, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if changes.PhoneNumber != nil {
		updates["phone_number"] = *changes.PhoneNumber
	}
	if changes.FirstName != nil {
		updates["first_name"] = *changes.FirstName
	}
	if changes.LastName != nil {
		updates["last_name"] = *changes.LastName
	}
	if changes.Gender != nil {
		updates["gender"] = *changes.Gender
	}
	if changes.IsActive != nil {
		updates["is_active"] = *changes.IsActive
	}
	if changes.TelegramUsername != nil {
		updates["telegram_username"] = *changes.TelegramUsername
	}
	if changes.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*changes.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		return &doctor, nil
	}

	if err := s.db.Model(&doctor).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&doctor, "id = ?", doctor.ID).Error; err != nil {
		return nil, err
	}

	return &doctor, nil
}

func (s *UserService) DeleteDoctor(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("doctor_id = ?", id).Delete(&models.OrganizationDoctor{}).Error; err != nil {
			return err
		}

		result := tx.Delete(&models.User{}, "id = ? AND role = ?", id, enums.UserRoleDoctor)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (s *UserService) createAndAssignDoctor(doctor *models.User, organizationID uuid.UUID) (*models.User, *models.OrganizationDoctor, error) {
	var assignment *models.OrganizationDoctor

	err := s.db.Transaction(func(tx *gorm.DB) error {
		tx.Create(doctor)

		if err := tx.First(&models.User{}, "id = ? AND role = ?", doctor.ID, enums.UserRoleDoctor).Error; err != nil {
			return err
		}

		if err := tx.First(&models.Organization{}, "id = ?", organizationID).Error; err != nil {
			return err
		}

		var existing models.OrganizationDoctor
		err := tx.Where("doctor_id = ? AND organization_id = ?", doctor.ID, organizationID).First(&existing).Error
		switch {
		case err == nil:
			existing.IsActive = true
			existing.LeftAt = nil
			if existing.JoinedAt.IsZero() {
				existing.JoinedAt = time.Now()
			}
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			assignment = &existing
			return nil
		case errors.Is(err, gorm.ErrRecordNotFound):
			newAssignment := models.OrganizationDoctor{
				DoctorID:       doctor.ID,
				OrganizationID: organizationID,
				JoinedAt:       time.Now(),
				IsActive:       true,
			}
			if err := tx.Create(&newAssignment).Error; err != nil {
				return err
			}
			assignment = &newAssignment
			return nil
		default:
			return err
		}
	})

	return doctor, assignment, err
}

func (s *UserService) UnassignFromOrganization(doctorID, organizationID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.OrganizationDoctor{}).
			Where("doctor_id = ? AND organization_id = ? AND is_active = ?", doctorID, organizationID, true).
			Updates(map[string]interface{}{
				"is_active": false,
				"left_at":   time.Now(),
			})

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (s *UserService) ListDoctorOrganizations(doctorID uuid.UUID, includeInactive bool) ([]models.OrganizationDoctor, error) {
	var relations []models.OrganizationDoctor
	query := s.db.Preload("Organization").Where("doctor_id = ?", doctorID)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

// Patient flows

type CreatePatientUserInput struct {
	PhoneNumber      string
	Password         string
	FirstName        string
	LastName         string
	Gender           *enums.Gender
	IsActive         *bool
	TelegramUsername string
}

func (s *UserService) CreatePatientUser(input CreatePatientUserInput) (*models.User, error) {
	user := models.User{
		PhoneNumber:      input.PhoneNumber,
		PasswordHash:     input.Password,
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Gender:           input.Gender,
		Role:             enums.UserRolePatient,
		TelegramUsername: input.TelegramUsername,
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

type UpdatePatientUserInput struct {
	Email            *string
	PhoneNumber      *string
	FirstName        *string
	LastName         *string
	Gender           *enums.Gender
	IsActive         *bool
	Password         *string
	TelegramUsername *string
}

func (s *UserService) UpdatePatientUser(id uuid.UUID, input UpdatePatientUserInput) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ? AND role = ?", id, enums.UserRolePatient).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.PhoneNumber != nil {
		updates["phone_number"] = *input.PhoneNumber
	}
	if input.FirstName != nil {
		updates["first_name"] = *input.FirstName
	}
	if input.LastName != nil {
		updates["last_name"] = *input.LastName
	}
	if input.Gender != nil {
		updates["gender"] = *input.Gender
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.TelegramUsername != nil {
		updates["telegram_username"] = *input.TelegramUsername
	}
	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		updates["password_hash"] = string(hashedPassword)
	}

	if len(updates) == 0 {
		return &user, nil
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&user, "id = ?", user.ID).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

type PatientMedicalInput struct {
	DoctorID                 uuid.UUID
	ConditionSummary         string
	Comorbidities            []string
	CurrentMedications       models.JSONB
	Allergies                []string
	BaselineVitals           models.JSONB
	RiskLevel                *enums.RiskLevel
	MonitoringFrequency      *enums.MonitoringFrequency
	Status                   *enums.PatientStatus
	DischargeDate            *time.Time
	DischargeNotes           *string
	EmergencyContactName     *string
	EmergencyContactPhone    *string
	EmergencyContactRelation *string
}

func (s *UserService) CreatePatientMedicalInfo(userID uuid.UUID, input PatientMedicalInput) (*models.Patient, error) {
	var patient models.Patient

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// ensure user exists and is patient
		if err := tx.First(&models.User{}, "id = ? AND role = ?", userID, enums.UserRolePatient).Error; err != nil {
			return err
		}

		// ensure doctor exists
		if err := tx.First(&models.User{}, "id = ? AND role = ?", input.DoctorID, enums.UserRoleDoctor).Error; err != nil {
			return err
		}

		// prevent duplicate patient record
		var existing models.Patient
		if err := tx.First(&existing, "user_id = ?", userID).Error; err == nil {
			return errors.New("patient medical info already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		patient = models.Patient{
			UserID:                   userID,
			DoctorID:                 input.DoctorID,
			ConditionSummary:         input.ConditionSummary,
			Comorbidities:            models.StringArray(input.Comorbidities),
			CurrentMedications:       input.CurrentMedications,
			Allergies:                models.StringArray(input.Allergies),
			BaselineVitals:           input.BaselineVitals,
			DischargeDate:            input.DischargeDate,
			DischargeNotes:           input.DischargeNotes,
			EmergencyContactName:     input.EmergencyContactName,
			EmergencyContactPhone:    input.EmergencyContactPhone,
			EmergencyContactRelation: input.EmergencyContactRelation,
		}

		if input.RiskLevel != nil {
			patient.RiskLevel = *input.RiskLevel
		} else {
			patient.RiskLevel = enums.RiskLevelMedium
		}
		if input.MonitoringFrequency != nil {
			patient.MonitoringFrequency = *input.MonitoringFrequency
		} else {
			patient.MonitoringFrequency = enums.MonitoringFrequencyDaily
		}
		if input.Status != nil {
			patient.Status = *input.Status
		} else {
			patient.Status = enums.PatientStatusActive
		}

		if err := tx.Create(&patient).Error; err != nil {
			return err
		}

		return tx.Preload("User").Preload("Doctor").First(&patient, "id = ?", patient.ID).Error
	})

	return &patient, err
}

type PatientMedicalUpdate struct {
	DoctorID                 *uuid.UUID
	ConditionSummary         *string
	Comorbidities            *[]string
	CurrentMedications       *models.JSONB
	Allergies                *[]string
	BaselineVitals           *models.JSONB
	RiskLevel                *enums.RiskLevel
	MonitoringFrequency      *enums.MonitoringFrequency
	Status                   *enums.PatientStatus
	DischargeDate            *time.Time
	DischargeNotes           *string
	EmergencyContactName     *string
	EmergencyContactPhone    *string
	EmergencyContactRelation *string
}

func (s *UserService) UpdatePatientMedicalInfo(userID uuid.UUID, input PatientMedicalUpdate) (*models.Patient, error) {
	var patient models.Patient
	if err := s.db.First(&patient, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if input.DoctorID != nil {
		if err := s.db.First(&models.User{}, "id = ? AND role = ?", *input.DoctorID, enums.UserRoleDoctor).Error; err != nil {
			return nil, err
		}
		updates["doctor_id"] = *input.DoctorID
	}
	if input.ConditionSummary != nil {
		updates["condition_summary"] = *input.ConditionSummary
	}
	if input.Comorbidities != nil {
		updates["comorbidities"] = models.StringArray(*input.Comorbidities)
	}
	if input.CurrentMedications != nil {
		updates["current_medications"] = *input.CurrentMedications
	}
	if input.Allergies != nil {
		updates["allergies"] = models.StringArray(*input.Allergies)
	}
	if input.BaselineVitals != nil {
		updates["baseline_vitals"] = *input.BaselineVitals
	}
	if input.RiskLevel != nil {
		updates["risk_level"] = *input.RiskLevel
	}
	if input.MonitoringFrequency != nil {
		updates["monitoring_frequency"] = *input.MonitoringFrequency
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.DischargeDate != nil {
		updates["discharge_date"] = input.DischargeDate
	}
	if input.DischargeNotes != nil {
		updates["discharge_notes"] = input.DischargeNotes
	}
	if input.EmergencyContactName != nil {
		updates["emergency_contact_name"] = input.EmergencyContactName
	}
	if input.EmergencyContactPhone != nil {
		updates["emergency_contact_phone"] = input.EmergencyContactPhone
	}
	if input.EmergencyContactRelation != nil {
		updates["emergency_contact_relation"] = input.EmergencyContactRelation
	}

	if len(updates) == 0 {
		return &patient, nil
	}

	if err := s.db.Model(&patient).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("User").Preload("Doctor").First(&patient, "id = ?", patient.ID).Error; err != nil {
		return nil, err
	}

	return &patient, nil
}

func (s *UserService) GetPatientDetailsByUserID(userID uuid.UUID) (*models.Patient, error) {
	var patient models.Patient
	if err := s.db.Preload("User").Preload("Doctor").First(&patient, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (s *UserService) ListPatientUsers(includeInactive bool) ([]models.User, error) {
	var patients []models.User
	query := s.db.Where("role = ?", enums.UserRolePatient)
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Order("created_at DESC").Find(&patients).Error; err != nil {
		return nil, err
	}
	return patients, nil
}

func (s *UserService) GetUserByTelegramUsername(username string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "telegram_username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type PatientCompleteData struct {
	User          *models.User          `json:"user"`
	Patient       *models.Patient       `json:"patient"`
	Checkins      []models.Checkin      `json:"checkins"`
	VitalReadings []models.VitalReading `json:"vital_readings"`
}

func (s *UserService) GetPatientCompleteData(userID uuid.UUID) (*PatientCompleteData, error) {
	var user models.User
	if err := s.db.First(&user, "id = ? AND role = ?", userID, enums.UserRolePatient).Error; err != nil {
		return nil, err
	}

	var patient models.Patient
	if err := s.db.Preload("Doctor").First(&patient, "user_id = ?", userID).Error; err != nil {
		return nil, err
	}

	var checkins []models.Checkin
	if err := s.db.Where("patient_id = ?", patient.ID).
		Order("created_at DESC").
		Find(&checkins).Error; err != nil {
		return nil, err
	}

	var vitals []models.VitalReading
	if err := s.db.Where("patient_id = ?", patient.ID).
		Order("created_at DESC").
		Find(&vitals).Error; err != nil {
		return nil, err
	}

	return &PatientCompleteData{
		User:          &user,
		Patient:       &patient,
		Checkins:      checkins,
		VitalReadings: vitals,
	}, nil
}

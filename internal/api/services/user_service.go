package services

import (
	"errors"
	"log/slog"
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/google/uuid"
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
	PhoneNumber *string
	FirstName   *string
	LastName    *string
	Gender      *enums.Gender
	IsActive    *bool
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

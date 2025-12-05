package services

import (
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationService struct {
	db *gorm.DB
}

func NewOrganizationService(db *gorm.DB) *OrganizationService {
	return &OrganizationService{db: db}
}

func (s *OrganizationService) Create(org *models.Organization) error {
	return s.db.Create(org).Error
}

func (s *OrganizationService) GetByID(id uuid.UUID) (*models.Organization, error) {
	var organization models.Organization
	if err := s.db.First(&organization, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &organization, nil
}

func (s *OrganizationService) List(includeInactive bool) ([]models.Organization, error) {
	var organizations []models.Organization
	query := s.db
	if !includeInactive {
		query = query.Where("is_active = ?", true)
	}
	if err := query.Order("created_at DESC").Find(&organizations).Error; err != nil {
		return nil, err
	}
	return organizations, nil
}

type OrganizationUpdate struct {
	Name          *string
	Address       *string
	LicenseNumber *string
	ContactEmail  *string
	ContactPhone  *string
	ManagerID     *uuid.UUID
	IsActive      *bool
}

func (s *OrganizationService) Update(id uuid.UUID, changes OrganizationUpdate) (*models.Organization, error) {
	var organization models.Organization
	if err := s.db.First(&organization, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if changes.Name != nil {
		updates["name"] = *changes.Name
	}
	if changes.Address != nil {
		updates["address"] = *changes.Address
	}
	if changes.LicenseNumber != nil {
		updates["license_number"] = *changes.LicenseNumber
	}
	if changes.ContactEmail != nil {
		updates["contact_email"] = *changes.ContactEmail
	}
	if changes.ContactPhone != nil {
		updates["contact_phone"] = *changes.ContactPhone
	}
	if changes.ManagerID != nil {
		updates["manager_id"] = *changes.ManagerID
	}
	if changes.IsActive != nil {
		updates["is_active"] = *changes.IsActive
	}

	if len(updates) == 0 {
		return &organization, nil
	}

	if err := s.db.Model(&organization).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&organization, "id = ?", organization.ID).Error; err != nil {
		return nil, err
	}

	return &organization, nil
}

func (s *OrganizationService) Delete(id uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("organization_id = ?", id).Delete(&models.OrganizationDoctor{}).Error; err != nil {
			return err
		}

		result := tx.Delete(&models.Organization{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

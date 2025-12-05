package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationDoctor struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	DoctorID       uuid.UUID  `gorm:"column:doctor_id;type:uuid;not null;index:idx_org_doctors_composite"`
	OrganizationID uuid.UUID  `gorm:"column:organization_id;type:uuid;not null;index:idx_org_doctors_composite"`
	JoinedAt       time.Time  `gorm:"column:joined_at;type:timestamptz;default:now()"`
	LeftAt         *time.Time `gorm:"column:left_at;type:timestamptz"`
	IsActive       bool       `gorm:"column:is_active;default:true"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;type:timestamptz;default:now()"`

	Doctor       *User         `gorm:"foreignKey:DoctorID"`
	Organization *Organization `gorm:"foreignKey:OrganizationID"`
}

func (od *OrganizationDoctor) BeforeCreate(tx *gorm.DB) error {
	if od.ID == uuid.Nil {
		od.ID = uuid.New()
	}
	return nil
}

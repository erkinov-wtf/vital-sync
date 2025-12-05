package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name          string    `gorm:"column:name;type:varchar(255);not null"`
	Address       *string   `gorm:"column:address;type:text"`
	LicenseNumber string    `gorm:"column:license_number;type:varchar(100);not null;uniqueIndex"`
	ContactEmail  *string   `gorm:"column:contact_email;type:varchar(255)"`
	ContactPhone  *string   `gorm:"column:contact_phone;type:varchar(20)"`
	ManagerID     uuid.UUID `gorm:"column:manager_id;type:uuid;index"`
	IsActive      bool      `gorm:"column:is_active;default:true"`
	CreatedAt     time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`

	Manager *User `gorm:"foreignKey:ManagerID"`
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

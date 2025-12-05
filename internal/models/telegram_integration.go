package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TelegramIntegration struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	PatientID        uuid.UUID  `gorm:"column:patient_id;type:uuid;not null;uniqueIndex"`
	TelegramChatID   *int64     `gorm:"column:telegram_chat_id;type:bigint;uniqueIndex"`
	TelegramUsername *string    `gorm:"column:telegram_username;type:varchar(255)"`
	InviteToken      string     `gorm:"column:invite_token;type:varchar(100);uniqueIndex"`
	IsActivated      bool       `gorm:"column:is_activated;default:false"`
	ActivatedAt      *time.Time `gorm:"column:activated_at;type:timestamptz"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:timestamptz;default:now()"`

	Patient *Patient `gorm:"foreignKey:PatientID"`
}

func (ti *TelegramIntegration) BeforeCreate(tx *gorm.DB) error {
	if ti.ID == uuid.Nil {
		ti.ID = uuid.New()
	}
	return nil
}

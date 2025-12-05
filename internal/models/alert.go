package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Alert struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PatientID uuid.UUID  `gorm:"column:patient_id;type:uuid;not null;index"`
	CheckinID *uuid.UUID `gorm:"column:checkin_id;type:uuid"`

	Severity  string `gorm:"column:severity;type:varchar(20);not null;index"` // low, medium, high, critical
	AlertType string `gorm:"column:alert_type;type:varchar(50);not null"`     // vital_abnormal, no_response, sentiment_negative, pattern_detected

	Title   string `gorm:"column:title;type:varchar(255);not null"`
	Message string `gorm:"column:message;type:text;not null"`
	Details JSONB  `gorm:"column:details;type:jsonb"`

	// Acknowledgment
	IsAcknowledged bool       `gorm:"column:is_acknowledged;default:false;index"`
	AcknowledgedBy *uuid.UUID `gorm:"column:acknowledged_by;type:uuid"`
	AcknowledgedAt *time.Time `gorm:"column:acknowledged_at;type:timestamptz"`
	ActionTaken    *string    `gorm:"column:action_taken;type:text"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now();index:idx_alerts_created_at,sort:desc"`

	Patient      *Patient `gorm:"foreignKey:PatientID"`
	Checkin      *Checkin `gorm:"foreignKey:CheckinID"`
	Acknowledger *User    `gorm:"foreignKey:AcknowledgedBy"`
}

func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

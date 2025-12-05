package models

import (
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Checkin struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PatientID  uuid.UUID  `gorm:"column:patient_id;type:uuid;not null;index"`
	ScheduleID *uuid.UUID `gorm:"column:schedule_id;type:uuid"`

	// Check-in Flow
	Status      enums.CheckinStatus `gorm:"column:status;type:varchar(20);default:'pending';index"` // pending, in_progress, completed, failed, missed
	InitiatedAt time.Time           `gorm:"column:initiated_at;type:timestamptz;default:now()"`
	CompletedAt *time.Time          `gorm:"column:completed_at;type:timestamptz"`

	// AI-Generated Questions
	Questions JSONB `gorm:"column:questions;type:jsonb;not null;default:'[]'"`

	// Patient Responses
	Answers     JSONB       `gorm:"column:answers;type:jsonb;default:'[]'"`
	RawMessages StringArray `gorm:"column:raw_messages;type:text[]"`

	// AI Analysis
	AIAnalysis    JSONB                `gorm:"column:ai_analysis;type:jsonb"`
	MedicalStatus *enums.MedicalStatus `gorm:"column:medical_status;type:varchar(20);index"` // normal, concern, urgent, critical
	RiskScore     *int                 `gorm:"column:risk_score;type:integer"`               // 0-100 scale

	// Doctor Review
	ReviewedBy  *uuid.UUID `gorm:"column:reviewed_by;type:uuid"`
	ReviewedAt  *time.Time `gorm:"column:reviewed_at;type:timestamptz"`
	DoctorNotes *string    `gorm:"column:doctor_notes;type:text"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now();index:idx_checkins_created_at,sort:desc"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`

	Patient  *Patient         `gorm:"foreignKey:PatientID"`
	Schedule *CheckinSchedule `gorm:"foreignKey:ScheduleID"`
	Reviewer *User            `gorm:"foreignKey:ReviewedBy"`
}

func (c *Checkin) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

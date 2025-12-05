package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Patient struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID   uuid.UUID `gorm:"column:user_id;type:uuid;not null;uniqueIndex"`
	DoctorID uuid.UUID `gorm:"column:doctor_id;type:uuid;not null;index"`

	// Medical Information
	ConditionSummary   string      `gorm:"column:condition_summary;type:text;not null"`
	Comorbidities      StringArray `gorm:"column:comorbidities;type:text[]"`
	CurrentMedications JSONB       `gorm:"column:current_medications;type:jsonb;default:'[]'"`
	Allergies          StringArray `gorm:"column:allergies;type:text[]"`

	// Baseline Vitals
	BaselineVitals JSONB `gorm:"column:baseline_vitals;type:jsonb"`

	// Monitoring Configuration
	RiskLevel           string `gorm:"column:risk_level;type:varchar(20);default:'medium';index"`    // low, medium, high, critical
	MonitoringFrequency string `gorm:"column:monitoring_frequency;type:varchar(30);default:'daily'"` // twice_daily, daily, every_other_day, weekly

	// Status Tracking
	Status         string     `gorm:"column:status;type:varchar(20);default:'active';index"` // active, paused, discharged, critical
	DischargeDate  *time.Time `gorm:"column:discharge_date;type:timestamptz"`
	DischargeNotes *string    `gorm:"column:discharge_notes;type:text"`

	// Emergency Contact
	EmergencyContactName     *string `gorm:"column:emergency_contact_name;type:varchar(255)"`
	EmergencyContactPhone    *string `gorm:"column:emergency_contact_phone;type:varchar(20)"`
	EmergencyContactRelation *string `gorm:"column:emergency_contact_relation;type:varchar(50)"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;default:now()"`

	User   *User `gorm:"foreignKey:UserID"`
	Doctor *User `gorm:"foreignKey:DoctorID"`
}

func (p *Patient) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

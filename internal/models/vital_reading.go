package models

import (
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VitalReading struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CheckinID uuid.UUID `gorm:"column:checkin_id;type:uuid;not null;index"`
	PatientID uuid.UUID `gorm:"column:patient_id;type:uuid;not null;index"`

	VitalType    enums.VitalType  `gorm:"column:vital_type;type:varchar(50);not null;index"` // blood_pressure, glucose, heart_rate, temperature, weight, oxygen_saturation
	ValueNumeric *float64         `gorm:"column:value_numeric;type:decimal(10,2)"`
	ValueText    *string          `gorm:"column:value_text;type:varchar(50)"`
	Unit         *enums.VitalUnit `gorm:"column:unit;type:varchar(20)"` // mmHg, mg/dL, bpm, °C, kg, %

	IsAbnormal            bool     `gorm:"column:is_abnormal;default:false"`
	DeviationFromBaseline *float64 `gorm:"column:deviation_from_baseline;type:decimal(10,2)"`

	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;default:now();index:idx_vital_readings_created_at,sort:desc"`

	Checkin *Checkin `gorm:"foreignKey:CheckinID"`
	Patient *Patient `gorm:"foreignKey:PatientID"`
}

func (vr *VitalReading) BeforeCreate(tx *gorm.DB) error {
	if vr.ID == uuid.Nil {
		vr.ID = uuid.New()
	}
	return nil
}

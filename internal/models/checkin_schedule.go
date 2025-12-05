package models

import (
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CheckinSchedule struct {
	ID            uuid.UUID               `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PatientID     uuid.UUID               `gorm:"column:patient_id;type:uuid;not null;uniqueIndex"`
	Frequency     enums.ScheduleFrequency `gorm:"column:frequency;type:varchar(30);not null"` // twice_daily, daily, every_other_day, weekly
	TimeSlots     TimeArray               `gorm:"column:time_slots;type:time[]"`
	Timezone      string                  `gorm:"column:timezone;type:varchar(50);default:'Asia/Tashkent'"`
	IsActive      bool                    `gorm:"column:is_active;default:true"`
	NextCheckinAt *time.Time              `gorm:"column:next_checkin_at;type:timestamptz;index"`
	CreatedAt     time.Time               `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt     time.Time               `gorm:"column:updated_at;type:timestamptz;default:now()"`

	Patient *Patient `gorm:"foreignKey:PatientID"`
}

func (cs *CheckinSchedule) BeforeCreate(tx *gorm.DB) error {
	if cs.ID == uuid.Nil {
		cs.ID = uuid.New()
	}
	return nil
}

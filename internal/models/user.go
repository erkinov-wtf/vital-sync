package models

import (
	"time"

	"github.com/erkinov-wtf/vital-sync/internal/enums"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Email        string         `gorm:"column:email;type:varchar(255);not null;uniqueIndex"`
	PhoneNumber  string         `gorm:"column:phone_number;type:varchar(20);not null;uniqueIndex"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null"`
	FirstName    string         `gorm:"column:first_name;type:varchar(100);not null"`
	LastName     string         `gorm:"column:last_name;type:varchar(100);not null"`
	Role         enums.UserRole `gorm:"column:role;type:varchar(20);not null"` // admin, doctor, patient
	Gender       *enums.Gender  `gorm:"column:gender;type:varchar(10)"`        // male, female, other
	IsActive     bool           `gorm:"column:is_active;default:true"`
	LastLoginAt  *time.Time     `gorm:"column:last_login_at;type:timestamptz"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamptz;default:now()"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamptz;default:now()"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)

	return nil
}

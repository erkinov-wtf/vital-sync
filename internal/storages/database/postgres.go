package database

import (
	"fmt"
	"log/slog"

	"github.com/erkinov-wtf/vital-sync/internal/config"
	"github.com/erkinov-wtf/vital-sync/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDB struct {
	DB *gorm.DB
}

func LoadDB(cfg *config.Config, logger *slog.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%v search_path=%s TimeZone=%s sslmode=disable",
		cfg.Internal.Database.Host,
		cfg.Internal.Database.User,
		cfg.Internal.Database.Password,
		cfg.Internal.Database.Name,
		cfg.Internal.Database.Port,
		cfg.Internal.Database.Schema,
		cfg.Internal.Database.Timezone,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	logger.Info("database connected successfully")

	var allModels = []interface{}{
		&models.Alert{},
		&models.Checkin{},
		&models.CheckinSchedule{},
		&models.Organization{},
		&models.OrganizationDoctor{},
		&models.Patient{},
		&models.TelegramIntegration{},
		&models.User{},
		&models.VitalReading{},
	}

	if len(allModels) > 0 {
		if err := db.AutoMigrate(allModels...); err != nil {
			return nil, fmt.Errorf("auto migration failed: %v", err)
		}
		logger.Info("auto migration completed")
	}

	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

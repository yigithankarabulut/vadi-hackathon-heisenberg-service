package postgres

import (
	"fmt"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds PostgreSQL connection configuration
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewClient creates a new GORM PostgreSQL client
func NewClient(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logging.Info("PostgreSQL connection established",
		zap.String("host", config.Host),
		zap.String("port", config.Port),
		zap.String("dbname", config.DBName),
	)

	return db, nil
}

// AutoMigrate runs GORM AutoMigrate for all models
func AutoMigrate(db *gorm.DB) error {
	logging.Info("Running database migrations")

	if err := db.AutoMigrate(
		&model.Aircraft{},
		&model.Threshold{},
		&model.Geofence{},
		&model.Telemetry{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logging.Info("Database migrations completed successfully")

	return nil
}

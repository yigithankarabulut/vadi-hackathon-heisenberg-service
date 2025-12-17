package repository

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"gorm.io/gorm"
)

// TelemetryRepository defines telemetry repository operations
type TelemetryRepository interface {
	Create(telemetry *model.Telemetry) error
	CreateBatch(telemetries []*model.Telemetry) error
}

type telemetryRepository struct {
	db *gorm.DB
}

// NewTelemetryRepository creates a new telemetry repository
func NewTelemetryRepository(db *gorm.DB) TelemetryRepository {
	return &telemetryRepository{db: db}
}

// Create creates a new telemetry record
func (r *telemetryRepository) Create(telemetry *model.Telemetry) error {
	return r.db.Create(telemetry).Error
}

// CreateBatch creates multiple telemetry records in a batch
func (r *telemetryRepository) CreateBatch(telemetries []*model.Telemetry) error {
	if len(telemetries) == 0 {
		return nil
	}
	return r.db.CreateInBatches(telemetries, 100).Error
}

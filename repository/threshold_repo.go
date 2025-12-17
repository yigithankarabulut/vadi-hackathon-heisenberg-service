package repository

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"gorm.io/gorm"
)

// ThresholdRepository defines threshold repository operations
type ThresholdRepository interface {
	GetByAircraftID(aircraftID uint) ([]*model.Threshold, error)
	GetDefaults() ([]*model.Threshold, error)
	GetByAircraftIDAndMetric(aircraftID uint, metricName string) (*model.Threshold, error)
}

type thresholdRepository struct {
	db *gorm.DB
}

// NewThresholdRepository creates a new threshold repository
func NewThresholdRepository(db *gorm.DB) ThresholdRepository {
	return &thresholdRepository{db: db}
}

// GetByAircraftID retrieves all thresholds for a specific aircraft
func (r *thresholdRepository) GetByAircraftID(aircraftID uint) ([]*model.Threshold, error) {
	var thresholds []*model.Threshold
	if err := r.db.Where("aircraft_id = ?", aircraftID).Find(&thresholds).Error; err != nil {
		return nil, err
	}
	return thresholds, nil
}

// GetDefaults retrieves all default (global) thresholds
func (r *thresholdRepository) GetDefaults() ([]*model.Threshold, error) {
	var thresholds []*model.Threshold
	if err := r.db.Where("aircraft_id IS NULL AND is_default = ?", true).Find(&thresholds).Error; err != nil {
		return nil, err
	}
	return thresholds, nil
}

// GetByAircraftIDAndMetric retrieves a specific threshold for an aircraft and metric
// Falls back to default if aircraft-specific threshold doesn't exist
func (r *thresholdRepository) GetByAircraftIDAndMetric(aircraftID uint, metricName string) (*model.Threshold, error) {
	var threshold model.Threshold
	
	// First try to get aircraft-specific threshold
	err := r.db.Where("aircraft_id = ? AND metric_name = ?", aircraftID, metricName).First(&threshold).Error
	if err == nil {
		return &threshold, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Fall back to default threshold
	err = r.db.Where("aircraft_id IS NULL AND metric_name = ? AND is_default = ?", metricName, true).First(&threshold).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &threshold, nil
}


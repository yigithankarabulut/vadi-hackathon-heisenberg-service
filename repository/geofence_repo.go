package repository

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"gorm.io/gorm"
)

// GeofenceRepository defines geofence repository operations
type GeofenceRepository interface {
	GetAllActive() ([]*model.Geofence, error)
}

type geofenceRepository struct {
	db *gorm.DB
}

// NewGeofenceRepository creates a new geofence repository
func NewGeofenceRepository(db *gorm.DB) GeofenceRepository {
	return &geofenceRepository{db: db}
}

// GetAllActive retrieves all active geofences
func (r *geofenceRepository) GetAllActive() ([]*model.Geofence, error) {
	var geofences []*model.Geofence
	if err := r.db.Where("is_active = ?", true).Find(&geofences).Error; err != nil {
		return nil, err
	}
	return geofences, nil
}

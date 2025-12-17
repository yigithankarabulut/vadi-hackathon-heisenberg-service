package repository

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"gorm.io/gorm"
)

// AircraftRepository defines aircraft repository operations
type AircraftRepository interface {
	GetByMACAddress(macAddress string) (*model.Aircraft, error)
	GetByID(id uint) (*model.Aircraft, error)
}

type aircraftRepository struct {
	db *gorm.DB
}

// NewAircraftRepository creates a new aircraft repository
func NewAircraftRepository(db *gorm.DB) AircraftRepository {
	return &aircraftRepository{db: db}
}

// GetByMACAddress retrieves an aircraft by MAC address
func (r *aircraftRepository) GetByMACAddress(macAddress string) (*model.Aircraft, error) {
	var aircraft model.Aircraft
	if err := r.db.Where("mac_address = ?", macAddress).First(&aircraft).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &aircraft, nil
}

// GetByID retrieves an aircraft by ID
func (r *aircraftRepository) GetByID(id uint) (*model.Aircraft, error) {
	var aircraft model.Aircraft
	if err := r.db.First(&aircraft, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &aircraft, nil
}

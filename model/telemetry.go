package model

import (
	"time"

	"gorm.io/gorm"
)

// Telemetry represents processed telemetry data stored in TimescaleDB
type Telemetry struct {
	Time        time.Time `gorm:"primaryKey;type:timestamptz;not null" json:"time"`
	AircraftID  uint      `gorm:"index;not null" json:"aircraft_id"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Altitude    float64   `json:"altitude"`
	GroundSpeed float64   `json:"ground_speed"`
	Heading     float64   `json:"heading"`
	ClimbRate   float64   `json:"climb_rate"`
	Temperature *float64  `json:"temperature,omitempty"` // Optional, for future use
	HasAnomaly  bool      `gorm:"default:false" json:"has_anomaly"`
	AnomalyType string    `json:"anomaly_type,omitempty"` // threshold, geofence, both
	CreatedAt   time.Time `json:"created_at"`
}

// TableName specifies the table name for Telemetry
func (Telemetry) TableName() string {
	return "telemetry_data"
}

// BeforeCreate hook to set CreatedAt
func (t *Telemetry) BeforeCreate(tx *gorm.DB) error {
	if t.Time.IsZero() {
		t.Time = time.Now()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	return nil
}


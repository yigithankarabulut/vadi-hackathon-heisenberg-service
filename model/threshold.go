package model

import (
	"gorm.io/gorm"
)

// MetricName represents the type of metric being thresholded
type MetricName string

const (
	MetricGroundSpeed MetricName = "ground_speed"
	MetricAltitude    MetricName = "altitude"
	MetricClimbRate   MetricName = "climb_rate"
	MetricHeading     MetricName = "heading"
	MetricTemperature MetricName = "temperature"
)

// Threshold represents threshold values for telemetry metrics
type Threshold struct {
	gorm.Model
	AircraftID *uint    `gorm:"index" json:"aircraft_id,omitempty"` // NULL = global default
	MetricName string   `gorm:"not null;index" json:"metric_name"`  // ground_speed, altitude, etc.
	MaxValue   *float64 `json:"max_value,omitempty"`
	MinValue   *float64 `json:"min_value,omitempty"`
	IsDefault  bool     `gorm:"default:false" json:"is_default"`
}

// TableName specifies the table name for Threshold
func (Threshold) TableName() string {
	return "thresholds"
}

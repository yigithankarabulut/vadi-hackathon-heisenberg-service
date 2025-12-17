package model

import (
	"gorm.io/gorm"
)

// Geofence represents a restricted area (rectangle)
type Geofence struct {
	gorm.Model
	Name         string  `gorm:"not null" json:"name"`
	Description  string  `json:"description,omitempty"`
	MinLatitude  float64 `gorm:"not null" json:"min_latitude"`
	MaxLatitude  float64 `gorm:"not null" json:"max_latitude"`
	MinLongitude float64 `gorm:"not null" json:"min_longitude"`
	MaxLongitude float64 `gorm:"not null" json:"max_longitude"`
	IsActive     bool    `gorm:"default:true" json:"is_active"`
}

// TableName specifies the table name for Geofence
func (Geofence) TableName() string {
	return "geofences"
}

// ContainsPoint checks if a point (lat, lon) is inside the geofence rectangle
func (g *Geofence) ContainsPoint(lat, lon float64) bool {
	if !g.IsActive {
		return false
	}
	return lat >= g.MinLatitude && lat <= g.MaxLatitude &&
		lon >= g.MinLongitude && lon <= g.MaxLongitude
}

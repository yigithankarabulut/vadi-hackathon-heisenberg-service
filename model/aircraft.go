package model

import (
	"gorm.io/gorm"
)

// Aircraft represents an aircraft in the system
type Aircraft struct {
	gorm.Model
	MACAddress       string `gorm:"uniqueIndex;not null" json:"mac_address"`
	Name             string `gorm:"not null" json:"name"`
	CurrentAirportID *uint  `json:"current_airport_id,omitempty"`
	AssignedPilotID  *uint  `json:"assigned_pilot_id,omitempty"`
	OwnerID          uint   `gorm:"not null" json:"owner_id"`
	Status           string `gorm:"default:active" json:"status"`
}

// TableName specifies the table name for Aircraft
func (Aircraft) TableName() string {
	return "aircraft"
}

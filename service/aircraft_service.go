package service

import (
	"fmt"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
)

// AircraftService handles aircraft-related operations
type AircraftService interface {
	GetAircraftByMAC(macAddress string) (*model.Aircraft, error)
}

type aircraftService struct {
	aircraftRepo repository.AircraftRepository
}

// NewAircraftService creates a new aircraft service
func NewAircraftService(aircraftRepo repository.AircraftRepository) AircraftService {
	return &aircraftService{
		aircraftRepo: aircraftRepo,
	}
}

// GetAircraftByMAC retrieves an aircraft by MAC address
func (s *aircraftService) GetAircraftByMAC(macAddress string) (*model.Aircraft, error) {
	if macAddress == "" {
		return nil, fmt.Errorf("mac address cannot be empty")
	}

	aircraft, err := s.aircraftRepo.GetByMACAddress(macAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get aircraft by MAC: %w", err)
	}

	if aircraft == nil {
		return nil, fmt.Errorf("aircraft not found for MAC address: %s", macAddress)
	}

	return aircraft, nil
}

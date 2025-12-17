package service

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
)

// GeofenceService handles geofence checking operations
type GeofenceService interface {
	CheckGeofences(lat, lon float64) (bool, []*model.Geofence) // returns (isViolation, violatingGeofences)
}

type geofenceService struct {
	geofenceRepo repository.GeofenceRepository
}

// NewGeofenceService creates a new geofence service
func NewGeofenceService(geofenceRepo repository.GeofenceRepository) GeofenceService {
	return &geofenceService{
		geofenceRepo: geofenceRepo,
	}
}

// CheckGeofences checks if a point (lat, lon) is inside any active geofence
// Returns (isViolation, list of violating geofences)
func (s *geofenceService) CheckGeofences(lat, lon float64) (bool, []*model.Geofence) {
	geofences, err := s.geofenceRepo.GetAllActive()
	if err != nil {
		// Log error but return no violation
		return false, nil
	}

	var violatingGeofences []*model.Geofence
	for _, geofence := range geofences {
		if geofence.ContainsPoint(lat, lon) {
			violatingGeofences = append(violatingGeofences, geofence)
		}
	}

	return len(violatingGeofences) > 0, violatingGeofences
}

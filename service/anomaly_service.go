package service

import (
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
)

// AnomalyService handles anomaly detection by combining threshold and geofence checks
type AnomalyService interface {
	DetectAnomaly(aircraftID uint, telemetry *model.TelemetryDTO) *model.Anomaly
}

type anomalyService struct {
	thresholdService ThresholdService
	geofenceService  GeofenceService
}

// NewAnomalyService creates a new anomaly service
func NewAnomalyService(thresholdService ThresholdService, geofenceService GeofenceService) AnomalyService {
	return &anomalyService{
		thresholdService: thresholdService,
		geofenceService:  geofenceService,
	}
}

// DetectAnomaly detects anomalies by checking thresholds and geofences
func (s *anomalyService) DetectAnomaly(aircraftID uint, telemetry *model.TelemetryDTO) *model.Anomaly {
	// Check thresholds
	hasThresholdViolation, thresholdViolations := s.thresholdService.CheckThresholds(aircraftID, telemetry)

	// Check geofences
	hasGeofenceViolation, violatingGeofences := s.geofenceService.CheckGeofences(telemetry.Latitude, telemetry.Longitude)

	// Determine anomaly type
	var anomalyType model.AnomalyType
	var details string

	hasAnomaly := hasThresholdViolation || hasGeofenceViolation

	if !hasAnomaly {
		return &model.Anomaly{
			HasAnomaly: false,
		}
	}

	if hasThresholdViolation && hasGeofenceViolation {
		anomalyType = model.AnomalyTypeBoth
		details = "Threshold and geofence violations detected"
		if len(thresholdViolations) > 0 {
			details += ": " + thresholdViolations[0]
		}
		if len(violatingGeofences) > 0 {
			details += " - Inside geofence: " + violatingGeofences[0].Name
		}
	} else if hasThresholdViolation {
		anomalyType = model.AnomalyTypeThreshold
		if len(thresholdViolations) > 0 {
			details = thresholdViolations[0]
		}
	} else if hasGeofenceViolation {
		anomalyType = model.AnomalyTypeGeofence
		if len(violatingGeofences) > 0 {
			details = "Inside restricted area: " + violatingGeofences[0].Name
		}
	}

	return &model.Anomaly{
		HasAnomaly:  hasAnomaly,
		AnomalyType: anomalyType,
		Details:     details,
	}
}

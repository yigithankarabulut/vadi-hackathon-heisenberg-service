package service

import (
	"fmt"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
	"go.uber.org/zap"
)

// ThresholdService handles threshold checking operations
type ThresholdService interface {
	CheckThresholds(aircraftID uint, telemetry *model.TelemetryDTO) (bool, []string) // returns (hasViolation, violations)
}

type thresholdService struct {
	thresholdRepo repository.ThresholdRepository
}

// NewThresholdService creates a new threshold service
func NewThresholdService(thresholdRepo repository.ThresholdRepository) ThresholdService {
	return &thresholdService{
		thresholdRepo: thresholdRepo,
	}
}

// CheckThresholds checks if telemetry values violate any thresholds
// Returns (hasViolation, list of violation descriptions)
func (s *thresholdService) CheckThresholds(aircraftID uint, telemetry *model.TelemetryDTO) (bool, []string) {
	var violations []string

	// Check each metric
	metrics := map[string]float64{
		string(model.MetricGroundSpeed): telemetry.GroundSpeed,
		string(model.MetricAltitude):    telemetry.Altitude,
		string(model.MetricClimbRate):   telemetry.ClimbRate,
		string(model.MetricHeading):     telemetry.Heading,
	}

	for metricName, value := range metrics {
		threshold, err := s.thresholdRepo.GetByAircraftIDAndMetric(aircraftID, metricName)
		if err != nil {
			logging.Error("Failed to get threshold", zap.Error(err), zap.String("metric_name", metricName), zap.Uint("aircraft_id", aircraftID))
			continue
		}

		if threshold == nil {
			logging.Warn("No threshold defined, skipping", zap.String("metric_name", metricName), zap.Uint("aircraft_id", aircraftID))
			continue
		}

		if threshold.MaxValue != nil && value > *threshold.MaxValue {
			violations = append(violations, fmt.Sprintf("%s exceeds maximum: %.2f > %.2f", metricName, value, *threshold.MaxValue))
		}

		if threshold.MinValue != nil && value < *threshold.MinValue {
			violations = append(violations, fmt.Sprintf("%s below minimum: %.2f < %.2f", metricName, value, *threshold.MinValue))
		}
	}

	return len(violations) > 0, violations
}

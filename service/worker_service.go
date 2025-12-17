package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/consumer"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/publisher"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
	"go.uber.org/zap"
)

// WorkerService is the main worker service that processes telemetry data
type WorkerService interface {
	Start(ctx context.Context) error
	Stop() error
}

type workerService struct {
	streamConsumer  consumer.StreamConsumer
	aircraftService AircraftService
	anomalyService  AnomalyService
	telemetryRepo   repository.TelemetryRepository
	feedPublisher   publisher.FeedPublisher
}

// NewWorkerService creates a new worker service
func NewWorkerService(
	streamConsumer consumer.StreamConsumer,
	aircraftService AircraftService,
	anomalyService AnomalyService,
	telemetryRepo repository.TelemetryRepository,
	feedPublisher publisher.FeedPublisher,
) WorkerService {
	return &workerService{
		streamConsumer:  streamConsumer,
		aircraftService: aircraftService,
		anomalyService:  anomalyService,
		telemetryRepo:   telemetryRepo,
		feedPublisher:   feedPublisher,
	}
}

// Start starts the worker service
func (w *workerService) Start(ctx context.Context) error {
	logging.Info("Starting worker service")

	return w.streamConsumer.Consume(ctx, func(entry *consumer.StreamEntry) error {
		return w.processEntry(ctx, entry)
	})
}

// Stop stops the worker service
func (w *workerService) Stop() error {
	logging.Info("Stopping worker service")
	return nil
}

// processEntry processes a single stream entry
func (w *workerService) processEntry(ctx context.Context, entry *consumer.StreamEntry) error {
	// Get aircraft by MAC address
	aircraft, err := w.aircraftService.GetAircraftByMAC(entry.PlaneID)
	if err != nil {
		logging.Warn("Aircraft not found, skipping entry",
			zap.String("mac_address", entry.PlaneID),
			zap.Error(err),
		)
		return nil // Skip entries for unknown aircraft
	}

	// Detect anomalies
	anomaly := w.anomalyService.DetectAnomaly(aircraft.ID, entry.Telemetry)

	// Convert timestamp
	timestamp := time.Unix(int64(entry.Telemetry.Timestamp), 0)
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Create telemetry record
	telemetry := &model.Telemetry{
		Time:        timestamp,
		AircraftID:  aircraft.ID,
		Latitude:    entry.Telemetry.Latitude,
		Longitude:   entry.Telemetry.Longitude,
		Altitude:    entry.Telemetry.Altitude,
		GroundSpeed: entry.Telemetry.GroundSpeed,
		Heading:     entry.Telemetry.Heading,
		ClimbRate:   entry.Telemetry.ClimbRate,
		HasAnomaly:  anomaly.HasAnomaly,
		AnomalyType: string(anomaly.AnomalyType),
	}

	// Save to database
	if err := w.telemetryRepo.Create(telemetry); err != nil {
		logging.Error("Failed to save telemetry to database",
			zap.Error(err),
			zap.Uint("aircraft_id", aircraft.ID),
			zap.String("anomaly_type", string(anomaly.AnomalyType)),
		)
		return fmt.Errorf("failed to save telemetry to database: %w", err)
	}

	// Publish to global feed (always)
	if err := w.feedPublisher.PublishGlobalTelemetry(ctx, telemetry); err != nil {
		logging.Error("Failed to publish to global feed",
			zap.Error(err),
			zap.Uint("aircraft_id", aircraft.ID),
			zap.String("anomaly_type", string(anomaly.AnomalyType)),
		)
		// Don't return error - continue processing
	}

	// Publish to alert feed if anomaly detected
	if anomaly.HasAnomaly {
		if err := w.feedPublisher.PublishAlert(ctx, telemetry, anomaly); err != nil {
			logging.Error("Failed to publish alert",
				zap.Error(err),
				zap.Uint("aircraft_id", aircraft.ID),
				zap.String("anomaly_type", string(anomaly.AnomalyType)),
			)
			// Don't return error - continue processing
		}
	}

	logging.Debug("Processed telemetry entry",
		zap.Uint("aircraft_id", aircraft.ID),
		zap.Bool("has_anomaly", anomaly.HasAnomaly),
		zap.String("anomaly_type", string(anomaly.AnomalyType)),
	)

	return nil
}

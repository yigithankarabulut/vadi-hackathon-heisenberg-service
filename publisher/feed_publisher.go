package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/redis"
	"go.uber.org/zap"
)

// FeedPublisher handles publishing to Redis Pub/Sub channels
type FeedPublisher interface {
	PublishGlobalTelemetry(ctx context.Context, telemetry *model.Telemetry) error
	PublishAlert(ctx context.Context, telemetry *model.Telemetry, anomaly *model.Anomaly) error
}

type feedPublisher struct {
	redisClient       redis.Client
	globalFeedChannel string
	alertFeedChannel  string
}

// NewFeedPublisher creates a new feed publisher
func NewFeedPublisher(redisClient redis.Client, globalFeedChannel, alertFeedChannel string) FeedPublisher {
	return &feedPublisher{
		redisClient:       redisClient,
		globalFeedChannel: globalFeedChannel,
		alertFeedChannel:  alertFeedChannel,
	}
}

// PublishGlobalTelemetry publishes processed telemetry to global_telemetry_feed
func (p *feedPublisher) PublishGlobalTelemetry(ctx context.Context, telemetry *model.Telemetry) error {
	data, err := json.Marshal(telemetry)
	if err != nil {
		return fmt.Errorf("failed to marshal global telemetry message: %w", err)
	}

	if err := p.redisClient.PublishToChannel(ctx, p.globalFeedChannel, data); err != nil {
		return fmt.Errorf("failed to publish to global feed: %w", err)
	}

	return nil
}

// PublishAlert publishes alert to alert_feed when anomaly is detected
func (p *feedPublisher) PublishAlert(ctx context.Context, telemetry *model.Telemetry, anomaly *model.Anomaly) error {
	alertData := map[string]interface{}{
		"telemetry": telemetry,
		"anomaly":   anomaly,
	}

	data, err := json.Marshal(alertData)
	if err != nil {
		return fmt.Errorf("failed to marshal alert message: %w", err)
	}

	if err := p.redisClient.PublishToChannel(ctx, p.alertFeedChannel, data); err != nil {
		return fmt.Errorf("failed to publish to alert feed: %w", err)
	}

	logging.Info("Alert published to alert feed",
		zap.Uint("aircraft_id", telemetry.AircraftID),
		zap.String("anomaly_type", string(anomaly.AnomalyType)),
	)

	return nil
}

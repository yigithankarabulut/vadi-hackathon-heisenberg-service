package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	redisv8 "github.com/go-redis/redis/v8"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/model"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/redis"
	"go.uber.org/zap"
)

// StreamConsumer handles reading from Redis Stream
type StreamConsumer interface {
	Consume(ctx context.Context, handler func(entry *StreamEntry) error) error
}

// StreamEntry represents a single entry from Redis Stream
type StreamEntry struct {
	ID         string
	PlaneID    string
	Telemetry  *model.TelemetryDTO
	ReceivedAt time.Time
}

type streamConsumer struct {
	redisClient      redis.Client
	streamKey        string
	groupName        string
	consumerName     string
	wg               sync.WaitGroup
	mu               sync.Mutex
	activeGoroutines int64 // Track active goroutines for monitoring
}

// NewStreamConsumer creates a new stream consumer
func NewStreamConsumer(redisClient redis.Client, streamKey, groupName, consumerName string) StreamConsumer {
	// workerCount parameter is kept for API compatibility but not used
	// Each message is processed in its own goroutine
	return &streamConsumer{
		redisClient:  redisClient,
		streamKey:    streamKey,
		groupName:    groupName,
		consumerName: consumerName,
	}
}

// Consume reads from Redis Stream and processes each message in a separate goroutine
func (c *streamConsumer) Consume(ctx context.Context, handler func(entry *StreamEntry) error) error {
	// Create consumer group if it doesn't exist
	rdb := c.redisClient.GetRawClient()
	_, err := rdb.XGroupCreateMkStream(ctx, c.streamKey, c.groupName, "0").Result()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logging.Warn("Failed to create consumer group (may already exist)", zap.Error(err))
	}

	logging.Info("Stream consumer started (each message processed in separate goroutine)")

	// Main loop: Read from stream and process each message concurrently
	for {
		select {
		case <-ctx.Done():
			// Wait for all active goroutines to complete
			logging.Info("Waiting for active goroutines to complete", zap.Int64("active", c.getActiveGoroutines()))
			c.wg.Wait()
			return ctx.Err()
		default:
			// Read from stream
			streams, err := c.redisClient.ReadFromStream(ctx, c.streamKey, c.groupName, c.consumerName, 10)
			if err != nil {
				if err.Error() == "redis: nil" {
					// No data available, continue
					time.Sleep(100 * time.Millisecond)
					continue
				}
				if !errors.Is(err, context.Canceled) {
					logging.Error("Failed to read from stream", zap.Error(err))
					continue
				}
				time.Sleep(1 * time.Second)
				continue
			}

			// Process each message in a separate goroutine
			if len(streams) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					entry, err := c.parseMessage(message)
					if err != nil {
						logging.Error("Failed to parse stream message",
							zap.Error(err),
							zap.String("id", message.ID),
						)
						// Acknowledge even if parsing failed to avoid reprocessing
						if ackErr := c.redisClient.AcknowledgeStream(ctx, c.streamKey, c.groupName, message.ID); ackErr != nil {
							logging.Error("Failed to acknowledge invalid message",
								zap.Error(ackErr),
								zap.String("id", message.ID),
							)
						}
						continue
					}

					// Process each message in its own goroutine
					c.wg.Add(1)
					c.incrementActiveGoroutines()

					go func(msg redisv8.XMessage, entry *StreamEntry) {
						defer c.wg.Done()
						defer c.decrementActiveGoroutines()

						// Call handler
						if err := handler(entry); err != nil {
							logging.Error("Handler failed",
								zap.Error(err),
								zap.String("id", msg.ID),
								zap.String("plane_id", entry.PlaneID),
							)
							// Don't acknowledge on handler error - message will be retried
							return
						}

						// Acknowledge successful processing
						if err := c.redisClient.AcknowledgeStream(ctx, c.streamKey, c.groupName, msg.ID); err != nil {
							logging.Error("Failed to acknowledge message",
								zap.Error(err),
								zap.String("id", msg.ID),
							)
						}
					}(message, entry)
				}
			}
		}
	}
}

// getActiveGoroutines returns the current number of active goroutines
func (c *streamConsumer) getActiveGoroutines() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.activeGoroutines
}

// incrementActiveGoroutines increments the active goroutine counter
func (c *streamConsumer) incrementActiveGoroutines() {
	c.mu.Lock()
	c.activeGoroutines++
	c.mu.Unlock()
}

// decrementActiveGoroutines decrements the active goroutine counter
func (c *streamConsumer) decrementActiveGoroutines() {
	c.mu.Lock()
	c.activeGoroutines--
	c.mu.Unlock()
}

// parseMessage parses a Redis stream message into StreamEntry
func (c *streamConsumer) parseMessage(message redisv8.XMessage) (*StreamEntry, error) {
	entry := &StreamEntry{
		ID: message.ID,
	}

	// Extract plane_id
	if planeID, ok := message.Values["plane_id"].(string); ok {
		entry.PlaneID = planeID
	} else {
		return nil, fmt.Errorf("missing or invalid plane_id")
	}

	// Extract data_json
	dataJSON, ok := message.Values["data_json"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid data_json")
	}

	// Parse telemetry data
	var telemetry model.TelemetryDTO
	if err := json.Unmarshal([]byte(dataJSON), &telemetry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal telemetry: %w", err)
	}
	entry.Telemetry = &telemetry

	// Extract received_at
	if receivedAtStr, ok := message.Values["received_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, receivedAtStr); err == nil {
			entry.ReceivedAt = t
		} else {
			entry.ReceivedAt = time.Now()
		}
	} else {
		entry.ReceivedAt = time.Now()
	}

	return entry, nil
}

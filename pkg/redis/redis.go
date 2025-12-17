// pkg/redis/redis.go
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"go.uber.org/zap"
)

// Client interface defines Redis operations
type Client interface {
	// WriteToStream writes data to Redis stream
	WriteToStream(ctx context.Context, streamKey string, data map[string]interface{}) error
	// ReadFromStream reads data from Redis stream using consumer group
	ReadFromStream(ctx context.Context, streamKey, groupName, consumerName string, count int64) ([]redis.XStream, error)
	// AcknowledgeStream acknowledges processed messages
	AcknowledgeStream(ctx context.Context, streamKey, groupName string, ids ...string) error
	// PublishToChannel publishes message to Redis Pub/Sub channel
	PublishToChannel(ctx context.Context, channel string, message interface{}) error
	// WriteToDiskBuffer writes data to disk buffer as fallback
	WriteToDiskBuffer(ctx context.Context, payload []byte) error
	// RecoverFromDisk recovers data from disk buffer to Redis
	RecoverFromDisk(ctx context.Context, streamKey string) error
	// Close closes the Redis connection
	Close() error
	// GetRawClient returns the underlying redis client for advanced operations
	GetRawClient() *redis.Client
}

// Config holds Redis client configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

// DiskBufferConfig holds disk buffer configuration
type DiskBufferConfig struct {
	MaxDiskSize  int64
	FailoverFile string
}

type redisClient struct {
	rdb    *redis.Client
	config DiskBufferConfig
}

// NewClient creates a new Redis client instance
func NewClient(ctx context.Context, config Config) (Client, error) {
	opts := &redis.Options{
		Addr: config.Addr,
		DB:   config.DB,
	}

	// Only set password if it's not empty
	if config.Password != "" {
		opts.Password = config.Password
	}

	rdb := redis.NewClient(opts)

	// Test connection
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}

	logging.Info("Redis connection established",
		zap.String("addr", config.Addr),
		zap.Int("db", config.DB),
	)

	return &redisClient{
		rdb:    rdb,
		config: DiskBufferConfig{},
	}, nil
}

// NewClientWithDiskBuffer creates a new Redis client with custom disk buffer config
func NewClientWithDiskBuffer(ctx context.Context, config Config, diskConfig DiskBufferConfig) (Client, error) {
	opts := &redis.Options{
		Addr: config.Addr,
		DB:   config.DB,
	}

	// Only set password if it's not empty
	if config.Password != "" {
		opts.Password = config.Password
	}

	rdb := redis.NewClient(opts)

	// Test connection
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis connection error: %w", err)
	}

	return &redisClient{
		rdb:    rdb,
		config: diskConfig,
	}, nil
}

// WriteToStream writes data to Redis stream
func (c *redisClient) WriteToStream(ctx context.Context, streamKey string, data map[string]interface{}) error {
	// Convert time.Time values to RFC3339 strings for Redis compatibility
	convertedData := make(map[string]interface{})
	for k, v := range data {
		switch val := v.(type) {
		case time.Time:
			convertedData[k] = val.Format(time.RFC3339)
		default:
			convertedData[k] = v
		}
	}

	_, err := c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		MaxLen: 100000, // Optional: limit stream length
		Values: convertedData,
	}).Result()

	if err != nil {
		return fmt.Errorf("redis stream write error: %w", err)
	}

	return nil
}

// WriteToDiskBuffer writes data to disk buffer as fallback
func (c *redisClient) WriteToDiskBuffer(ctx context.Context, payload []byte) error {
	file, err := os.OpenFile(c.config.FailoverFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open disk buffer file: %w", err)
	}
	defer file.Close()

	// Write payload with newline (JSONL format)
	if _, err := file.Write(append(payload, '\n')); err != nil {
		return fmt.Errorf("failed to write to disk buffer: %w", err)
	}

	return nil
}

// RecoverFromDisk recovers data from disk buffer to Redis
func (c *redisClient) RecoverFromDisk(ctx context.Context, streamKey string) error {
	file, err := os.Open(c.config.FailoverFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to recover
		}
		return fmt.Errorf("failed to open disk buffer file: %w", err)
	}
	defer file.Close()

	// Read file line by line and write to Redis
	// This is a simplified version - in production, you might want to use bufio.Scanner
	// For now, return not implemented as the original code did
	return errors.New("not implemented: RecoverFromDisk logic")
}

// ReadFromStream reads data from Redis stream using consumer group
func (c *redisClient) ReadFromStream(ctx context.Context, streamKey, groupName, consumerName string, count int64) ([]redis.XStream, error) {
	streams, err := c.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamKey, ">"},
		Count:    count,
		Block:    0, // Block indefinitely until data is available
	}).Result()

	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("redis stream read error: %w", err)
	}

	return streams, nil
}

// AcknowledgeStream acknowledges processed messages
func (c *redisClient) AcknowledgeStream(ctx context.Context, streamKey, groupName string, ids ...string) error {
	if len(ids) == 0 {
		return nil
	}

	_, err := c.rdb.XAck(ctx, streamKey, groupName, ids...).Result()
	if err != nil {
		return fmt.Errorf("redis stream ack error: %w", err)
	}

	return nil
}

// PublishToChannel publishes message to Redis Pub/Sub channel
func (c *redisClient) PublishToChannel(ctx context.Context, channel string, message interface{}) error {
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	case []byte:
		msg = string(v)
	default:
		// Try to marshal to JSON if it's not a string
		jsonBytes, err := json.Marshal(message)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
		msg = string(jsonBytes)
	}

	_, err := c.rdb.Publish(ctx, channel, msg).Result()
	if err != nil {
		return fmt.Errorf("redis pub/sub publish error: %w", err)
	}

	return nil
}

// GetRawClient returns the underlying redis client for advanced operations
func (c *redisClient) GetRawClient() *redis.Client {
	return c.rdb
}

// Close closes the Redis connection
func (c *redisClient) Close() error {
	if c.rdb != nil {
		return c.rdb.Close()
	}
	return nil
}

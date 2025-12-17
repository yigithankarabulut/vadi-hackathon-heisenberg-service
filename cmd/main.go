package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/consumer"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/postgres"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/redis"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/publisher"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/service"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	serverEnvironment := getEnv("SERVER_ENV", "production")
	os.Setenv("SERVER_ENV", serverEnvironment)
	logging.CreateLogger(
		logging.SetLogLevelString(
			getEnv("LOG_LEVEL", "info"),
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	redisConfig := redis.Config{
		Addr:     getEnv("REDIS_ADDR", "89.47.113.24:6379"),
		Password: getEnv("REDIS_PASSWORD", ""), // flightredispass
		DB:       0,
	}

	redisClient, err := redis.NewClient(ctx, redisConfig)
	if err != nil {
		logging.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize PostgreSQL client
	postgresConfig := postgres.Config{
		Host:     getEnv("POSTGRES_HOST", "89.47.113.24"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		User:     getEnv("POSTGRES_USER", "rbac_admin"),
		Password: getEnv("POSTGRES_PASSWORD", "rbac_pass"),
		DBName:   getEnv("POSTGRES_DB", "fleetdb"),
		SSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
	}

	db, err := postgres.NewClient(postgresConfig)
	if err != nil {
		logging.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	// AutoMigrate models
	autoMigrateEnabled := getEnv("AUTO_MIGRATE", "true")
	if autoMigrateEnabled == "true" {
		if err := postgres.AutoMigrate(db); err != nil {
			logging.Fatal("Failed to run migrations", zap.Error(err))
		}
	}

	// Initialize repositories
	aircraftRepo := repository.NewAircraftRepository(db)
	thresholdRepo := repository.NewThresholdRepository(db)
	geofenceRepo := repository.NewGeofenceRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	// Initialize services
	aircraftService := service.NewAircraftService(aircraftRepo)
	thresholdService := service.NewThresholdService(thresholdRepo)
	geofenceService := service.NewGeofenceService(geofenceRepo)
	anomalyService := service.NewAnomalyService(thresholdService, geofenceService)

	// Initialize consumer
	streamKey := getEnv("REDIS_STREAM_KEY", "telemetry_stream")
	consumerGroup := getEnv("REDIS_CONSUMER_GROUP", "heisenberg-workers")
	consumerName := getEnv("REDIS_CONSUMER_NAME", fmt.Sprintf("heisenberg-worker/%s", getHostname()))

	streamConsumer := consumer.NewStreamConsumer(redisClient, streamKey, consumerGroup, consumerName)

	// Initialize publisher
	globalFeedChannel := getEnv("REDIS_PUBSUB_GLOBAL_FEED", "global_telemetry_feed")
	alertFeedChannel := getEnv("REDIS_PUBSUB_ALERT_FEED", "alert_feed")
	feedPublisher := publisher.NewFeedPublisher(redisClient, globalFeedChannel, alertFeedChannel)

	// Initialize worker service
	workerService := service.NewWorkerService(
		streamConsumer,
		aircraftService,
		anomalyService,
		telemetryRepo,
		feedPublisher,
	)

	// Setup health check endpoint
	http.HandleFunc("/health", HealthCheckHandler)

	// Start HTTP server in a goroutine
	go func() {
		port := getEnv("SERVER_PORT", "1338")
		logging.Info("Health check server running", zap.String("port", port))
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			logging.Fatal("HTTP server error", zap.Error(err))
		}
	}()

	// Start worker service in a goroutine
	go func() {
		if err := workerService.Start(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				logging.Warn("Worker service received cancellation signal")
				return
			}
			logging.Fatal("Worker service started with error", zap.Error(err))
		}
	}()

	logging.Info("Heisenberg service started successfully")

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	<-sigChan
	logging.Warn("Shutdown signal received")

	// Stop the worker service gracefully
	cancel()
	workerService.Stop()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logging.Info("Service stopped gracefully")
}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

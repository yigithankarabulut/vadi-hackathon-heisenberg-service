package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/consumer"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/config"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/postgres"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/redis"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/publisher"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/repository"
	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/service"
	"go.uber.org/zap"
)

func main() {
	serverEnvironment := os.Getenv("SERVER_ENV")
	if serverEnvironment == "" {
		if err := os.Setenv("SERVER_ENV", "prod"); err != nil {
			log.Fatalf("Failed to set SERVER_ENV: %v", err)
		}
	}

	cfg, err := config.Load(serverEnvironment)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logging.CreateLogger(
		logging.SetLogLevelString(
			cfg.LogLevel,
		),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Redis client
	redisConfig := redis.Config{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	}

	redisClient, err := redis.NewClient(ctx, redisConfig)
	if err != nil {
		logging.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize PostgreSQL client
	postgresConfig := postgres.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		DBName:   cfg.PostgresDb,
		SSLMode:  cfg.PostgresSSLMode,
	}

	db, err := postgres.NewClient(postgresConfig)
	if err != nil {
		logging.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	// AutoMigrate models
	autoMigrateEnabled := cfg.AutoMigrate
	if autoMigrateEnabled {
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
	streamKey := cfg.RedisStreamKey
	consumerGroup := cfg.RedisConsumerGroup
	consumerName := fmt.Sprintf("heisenberg-worker/%s", getHostname())

	streamConsumer := consumer.NewStreamConsumer(redisClient, streamKey, consumerGroup, consumerName)

	// Initialize publisher
	globalFeedChannel := cfg.RedisPubSubGlobalFeed
	alertFeedChannel := cfg.RedisPubSubAlertFeed
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
		logging.Info("Health check server running", zap.String("port", cfg.Port))
		if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
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

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

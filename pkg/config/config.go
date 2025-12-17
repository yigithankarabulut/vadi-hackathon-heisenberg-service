package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/yigithankarabulut/vadi-hackathon-heisenberg-service/pkg/logging"
	"go.uber.org/zap"
)

var ErrEmptyServerEnv = errors.New("SERVER_ENV cannot be empty. Please provide a valid environment. local, prod, test etc")

const configFilePath = "./configs/appconfig.%s.json"

// Config is a struct that contains the config for the application.
type Config struct {
	AppName               string `json:"app_name"`
	Port                  string `json:"port"`
	SwaggerEnabled        bool   `json:"swagger_enabled"`
	LogLevel              string `json:"log_level"`
	RedisAddr             string `json:"redis_addr"`
	RedisPassword         string `json:"redis_password"`
	RedisStreamKey        string `json:"redis_stream_key"`
	RedisConsumerGroup    string `json:"redis_consumer_group"`
	RedisPubSubGlobalFeed string `json:"redis_pubsub_global_feed"`
	RedisPubSubAlertFeed  string `json:"redis_pubsub_alert_feed"`
	PostgresHost          string `json:"postgres_host"`
	PostgresPort          string `json:"postgres_port"`
	PostgresUser          string `json:"postgres_user"`
	PostgresPassword      string `json:"postgres_password"`
	PostgresDb            string `json:"postgres_db"`
	PostgresSSLMode       string `json:"postgres_sslmode"`
	AutoMigrate           bool   `json:"auto_migrate"`
}

// Load is a function that loads the config from the file.
func Load(serverEnv string) (*Config, error) {
	if serverEnv == "" {
		return nil, ErrEmptyServerEnv
	}

	file, err := os.Open(fmt.Sprintf(configFilePath, serverEnv))
	if err != nil {
		logging.Error("failed to open config file", zap.Error(err))

		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer func() { _ = file.Close() }()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		logging.Error("failed to decode config file", zap.Error(err))

		return nil, fmt.Errorf("failed to decode config file: %v", err)
	}

	return &config, nil
}

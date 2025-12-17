package logging

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	loggerOnce sync.Once // Singleton pattern for logger instance
	logger     *Logger   // Logger instance
)

// Logger is a struct that contains the logger and the logger config.
type Logger struct {
	logger       *zap.Logger
	loggerConfig LoggerConfig
}

// LoggerConfig is a struct that contains the logger config.
type LoggerConfig struct {
	Config           zap.Config
	ContextFieldFunc func(ctx context.Context) []zap.Field
}

// InfoCtx logs an info message with context fields.
func InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Info(msg, allFields...)
}

// ErrorCtx logs an error message with context fields.
func ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Error(msg, allFields...)
}

// WarnCtx logs a warning message with context fields.
func WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Warn(msg, allFields...)
}

// DebugCtx logs a debug message with context fields.
func DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Debug(msg, allFields...)
}

// PanicCtx logs a panic message with context fields.
func PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Panic(msg, allFields...)
}

// FatalCtx logs a fatal message with context fields.
func FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.Fatal(msg, allFields...)
}

// DPanicCtx logs a dpanic message with context fields.
func DPanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := GetContextFields(ctx)
	allFields := append(contextFields, fields...)
	logger.logger.DPanic(msg, allFields...)
}

// Info logs an info message.
func Info(msg string, fields ...zap.Field) {
	logger.logger.Info(msg, fields...)
}

// Error logs an error message.
func Error(msg string, fields ...zap.Field) {
	logger.logger.Error(msg, fields...)
}

// Warn logs a warning message.
func Warn(msg string, fields ...zap.Field) {
	logger.logger.Warn(msg, fields...)
}

// Debug logs a debug message.
func Debug(msg string, fields ...zap.Field) {
	logger.logger.Debug(msg, fields...)
}

// Fatal logs a fatal message.
func Fatal(msg string, fields ...zap.Field) {
	logger.logger.Fatal(msg, fields...)
}

// Panic logs a panic message.
func Panic(msg string, fields ...zap.Field) {
	logger.logger.Panic(msg, fields...)
}

// DPanic logs a dpanic message.
func DPanic(msg string, fields ...zap.Field) {
	logger.logger.DPanic(msg, fields...)
}

// NewLogger creates a new logger.
func NewLogger(config LoggerConfig) *Logger {
	config.Config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	zapLogger, _ := config.Config.Build(zap.AddCallerSkip(1))

	return &Logger{
		logger:       zapLogger,
		loggerConfig: config,
	}
}

// GetLogger returns the logger.
func GetLogger(config LoggerConfig) *Logger {
	loggerOnce.Do(func() {
		logger = NewLogger(config)
	})

	return logger
}

// GetContextFields returns the context fields.
func GetContextFields(ctx context.Context) []zap.Field {
	if ctx != nil && logger.loggerConfig.ContextFieldFunc != nil {
		return logger.loggerConfig.ContextFieldFunc(ctx)
	}

	return nil
}

// SetLogLevel sets the log level.
func SetLogLevel(logLevel zapcore.Level) {
	logger.loggerConfig.Config.Level.SetLevel(logLevel)
}

// SetLogLevelString sets the log level from a string.
func SetLogLevelString(logLevel string) zapcore.Level {
	logLevel = strings.ToLower(logLevel)
	switch logLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	case "panic":
		return zapcore.PanicLevel
	default:
		return zapcore.InfoLevel
	}
}

// CreateLogger creates a new logger with production config.
func CreateLogger(logLevel zapcore.Level) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)
	zapConfig.Sampling = nil
	zapConfig.EncoderConfig.MessageKey = "message"
	zapConfig.EncoderConfig.CallerKey = "caller"
	zapConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	logConfig := LoggerConfig{
		Config:           zapConfig,
		ContextFieldFunc: contextFields,
	}

	GetLogger(logConfig)
}

func contextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	serverEnvironment := os.Getenv("SERVER_ENV")
	fields = append(fields, zap.String("environment", serverEnvironment))

	return fields
}

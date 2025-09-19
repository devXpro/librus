package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// Initialize sets up the global logger based on LOG_LEVEL environment variable
func Initialize() {
	logLevel := getLogLevelFromEnv()

	// Always use human-readable format
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// getLogLevelFromEnv parses LOG_LEVEL environment variable
func getLogLevelFromEnv() zapcore.Level {
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))

	switch logLevelStr {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN", "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel // Default to INFO
	}
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Convenience functions for common logging patterns
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Convenience functions with error handling
func InfoWithError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Logger.Info(msg, allFields...)
}

func WarnWithError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Logger.Warn(msg, allFields...)
}

func ErrorWithError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Logger.Error(msg, allFields...)
}

func FatalWithError(msg string, err error, fields ...zap.Field) {
	allFields := append(fields, zap.Error(err))
	Logger.Fatal(msg, allFields...)
}

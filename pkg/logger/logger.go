package logger

import (
	"github.com/anatoly_dev/go-users/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Initialize(cfg *config.Config) error {
	var err error
	Logger, err = NewLogger(cfg)
	if err != nil {
		return err
	}
	return nil
}

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	environment := cfg.Environment

	if environment == "" {
		environment = "development"
	}

	devConfig := zap.NewDevelopmentConfig()
	devConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	prodConfig := zap.NewProductionConfig()
	prodConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var config zap.Config
	if environment == "production" {
		config = prodConfig
	} else {
		config = devConfig
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.OutputPaths = []string{"stdout"}
		config.EncoderConfig.MessageKey = "M"
		config.EncoderConfig.CallerKey = "C"
		config.EncoderConfig.LevelKey = "L"
		config.EncoderConfig.TimeKey = "T"
		config.Encoding = "console"
	}

	level := zap.InfoLevel
	if cfg.LogLevel != "" {
		err := level.UnmarshalText([]byte(cfg.LogLevel))
		if err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.WithOptions(zap.AddCallerSkip(1)), nil
}

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

func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

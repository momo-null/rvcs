package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"rvcs/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitDefaultLogger 初始化默认日志（用于配置加载失败时）
func InitDefaultLogger() {
	Logger, _ = zap.NewDevelopment()
}

// InitLogger 初始化日志
func InitLogger(cfg *config.Config) error {
	var zapConfig zap.Config

	if cfg.Log.Format == "json" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// 设置日志级别
	switch cfg.Log.Level {
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	// 设置编码器
	if cfg.Log.Format == "console" {
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// 设置输出
	if cfg.Log.Output == "file" {
		// 创建 logs 目录
		logDir := "./logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create logs directory: %w", err)
		}

		// 使用绝对路径
		absLogDir, err := filepath.Abs(logDir)
		if err == nil {
			zapConfig.OutputPaths = []string{filepath.Join(absLogDir, "app.log")}
			zapConfig.ErrorOutputPaths = []string{filepath.Join(absLogDir, "error.log")}
		} else {
			zapConfig.OutputPaths = []string{"./logs/app.log"}
			zapConfig.ErrorOutputPaths = []string{"./logs/error.log"}
		}
	}

	// 创建logger
	var err error
	Logger, err = zapConfig.Build()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	return nil
}

// Sync 同步日志
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	return Logger.With(zap.Any(key, value))
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) *zap.Logger {
	if Logger == nil {
		return zap.NewNop()
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return Logger.With(zapFields...)
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	} else {
		fmt.Printf("[DEBUG] %s\n", msg)
	}
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	} else {
		fmt.Printf("[INFO] %s\n", msg)
	}
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	} else {
		fmt.Printf("[WARN] %s\n", msg)
	}
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	} else {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", msg)
	}
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	} else {
		fmt.Fprintf(os.Stderr, "[FATAL] %s\n", msg)
		os.Exit(1)
	}
}

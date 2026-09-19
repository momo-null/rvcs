package database

import (
	"fmt"

	"rvcs/internal/config"
	"rvcs/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitMySQL initializes MySQL connection.
func InitMySQL(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	var logLevel logger.LogLevel
	switch cfg.Log.Level {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Warn
	case "warn":
		logLevel = logger.Error
	case "error":
		logLevel = logger.Silent
	default:
		logLevel = logger.Warn
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	DB = db
	return db, nil
}

// Migrate runs schema migrations.
func Migrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Device{},
		&model.DeviceSession{},
		&model.DeviceActivityLog{},
		&model.Stream{},
		&model.Alert{},
		&model.AlertRule{},
		&model.DeviceGroup{},
		&model.DeviceGroupMember{},
		&model.DeviceCommand{},
		&model.RegistrationCode{},
	)
}

// Close closes DB connection.
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

package database

import (
	"fmt"
	"os"
	"path/filepath"
	"rvcs/internal/config"
	"rvcs/internal/model"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitSQLite 初始化SQLite数据库（根据配置选择文件或内存模式）
func InitSQLite(cfg *config.Config) (*gorm.DB, error) {
	var dsn string
	var modeInfo string

	// 调试信息
	if cfg != nil {
		fmt.Printf("Database config mode: %s\n", cfg.Database.Mode)
	} else {
		fmt.Println("No config provided")
	}

	if cfg != nil && cfg.Database.Mode == "sqlite-memory" {
		// 内存模式
		dsn = ":memory:"
		modeInfo = "memory mode"
	} else {
		// 文件模式（默认）
		dataDir := "./data"
		filename := "rvcs.db"

		// 如果配置中有指定路径，则使用配置值
		if cfg != nil {
			if cfg.Database.SQLiteDataDir != "" {
				dataDir = cfg.Database.SQLiteDataDir
			}
			if cfg.Database.SQLiteFilename != "" {
				filename = cfg.Database.SQLiteFilename
			}
		}

		// 确保数据目录存在
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		dsn = filepath.Join(dataDir, filename)
		modeInfo = fmt.Sprintf("file mode: %s", dsn)
	}

	var logLevel logger.LogLevel
	if cfg != nil {
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
	} else {
		logLevel = logger.Warn
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
	}

	// SQLite不需要设置连接池参数
	fmt.Printf("Using SQLite database (%s)\n", modeInfo)

	DB = db
	return db, nil
}

// CreateDefaultAdminUser 创建默认管理员账户
func CreateDefaultAdminUser(db *gorm.DB) error {
	// 检查是否已存在管理员用户
	var existingUser model.User
	result := db.Where("username = ?", "admin").First(&existingUser)
	if result.Error == nil {
		fmt.Println("Admin user already exists")
		return nil
	}

	// 创建默认管理员账户
	password := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := &model.User{
		ID:           "admin-user-id",
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		Role:         "admin",
		Status:       "active",
	}

	if err := db.Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	fmt.Println("✅ Default admin user created!")
	fmt.Println("   Username: admin")
	fmt.Println("   Password: admin123")
	fmt.Println("   Email: admin@example.com")

	return nil
}

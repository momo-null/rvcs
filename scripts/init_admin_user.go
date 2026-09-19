package main

import (
	"fmt"
	"log"

	"rvcs/internal/config"
	"rvcs/internal/database"
)

func main() {
	fmt.Println("=== Initializing Default Admin User ===")

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := database.InitMySQL(cfg)
	if err != nil {
		fmt.Println("MySQL connection failed, using SQLite memory database...")
		db, err = database.InitSQLite(cfg)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
	} else {
		fmt.Println("Connected to MySQL database")
	}

	// 执行迁移
	if err := database.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 创建默认管理员账户
	if err := database.CreateDefaultAdminUser(db); err != nil {
		log.Fatalf("Failed to create default admin user: %v", err)
	}

	fmt.Println("\n✅ System is ready!")
	fmt.Println("👉 Frontend: http://localhost:23000")
	fmt.Println("👉 Backend API: http://localhost:28080")
	fmt.Println("👉 Login with admin/admin123")
}

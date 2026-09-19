package main

import (
	"fmt"
	"rvcs/internal/config"
	"rvcs/internal/logger"
)

func main() {
	fmt.Println("=== 日志配置测试 ===")

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("配置加载失败: %v\n", err)
		return
	}

	fmt.Printf("日志配置:\n")
	fmt.Printf("- Level: %s\n", cfg.Log.Level)
	fmt.Printf("- Format: %s\n", cfg.Log.Format)
	fmt.Printf("- Output: %s\n", cfg.Log.Output)

	// 初始化日志
	if err := logger.InitLogger(cfg); err != nil {
		fmt.Printf("日志初始化失败: %v\n", err)
		return
	}

	// 测试日志输出
	logger.Info("这是测试信息日志")
	logger.Warn("这是测试警告日志")
	logger.Debug("这是测试调试日志")

	fmt.Println("日志测试完成，请检查 logs/ 目录下的日志文件")
}

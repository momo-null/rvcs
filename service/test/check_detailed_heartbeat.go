package main

import (
	"fmt"
	"log"
	"rvcs/internal/database"
	"rvcs/internal/model"
	"time"
)

func main() {
	// 初始化数据库连接
	db, err := database.InitSQLite(nil)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	fmt.Println("=== 详细设备心跳检查 ===")

	// 查询所有设备
	var devices []model.Device
	result := db.Find(&devices)
	if result.Error != nil {
		log.Fatalf("Failed to query devices: %v", result.Error)
	}

	fmt.Printf("总共找到 %d 个设备:\n\n", len(devices))

	now := time.Now()
	onlineThreshold := now.Add(-5 * time.Minute) // 5分钟内有心跳算在线

	for i, device := range devices {
		fmt.Printf("--- 设备 #%d ---\n", i+1)
		fmt.Printf("设备ID: %s\n", device.ID)
		fmt.Printf("设备名称: %s\n", device.Name)
		fmt.Printf("设备状态: %s\n", device.Status)

		lastSeen := "从未在线"
		if device.LastSeen != nil {
			lastSeen = device.LastSeen.Format("2006-01-02 15:04:05")
			timeDiff := now.Sub(*device.LastSeen)

			if device.LastSeen.After(onlineThreshold) {
				fmt.Printf("最后在线时间: %s (在线)\n", lastSeen)
			} else {
				fmt.Printf("最后在线时间: %s (%v前)\n", lastSeen, timeDiff.Truncate(time.Second))
			}
		} else {
			fmt.Printf("最后在线时间: %s\n", lastSeen)
		}

		// 检查心跳日志
		var heartbeatLogs []model.DeviceActivityLog
		db.Where("device_id = ? AND action = ?", device.ID, "heartbeat").
			Order("created_at DESC").
			Limit(5).
			Find(&heartbeatLogs)

		if len(heartbeatLogs) > 0 {
			fmt.Printf("最近心跳记录 (%d条):\n", len(heartbeatLogs))
			for _, log := range heartbeatLogs {
				fmt.Printf("  - %s: %s\n", log.CreatedAt.Format("2006-01-02 15:04:05"), log.Details)
			}
		} else {
			fmt.Println("无心跳记录")
		}

		// 检查所有活动日志
		var allLogs []model.DeviceActivityLog
		db.Where("device_id = ?", device.ID).
			Order("created_at DESC").
			Limit(10).
			Find(&allLogs)

		if len(allLogs) > 0 {
			fmt.Printf("最近活动日志 (%d条):\n", len(allLogs))
			for _, log := range allLogs {
				fmt.Printf("  - [%s] %s: %s\n",
					log.CreatedAt.Format("2006-01-02 15:04:05"),
					log.Action,
					log.Details)
			}
		}

		fmt.Println()
	}

	// 统计信息
	fmt.Println("=== 统计信息 ===")

	// 在线设备统计
	var onlineCount int64
	db.Model(&model.Device{}).Where("status = ?", "online").Count(&onlineCount)

	var offlineCount int64
	db.Model(&model.Device{}).Where("status = ?", "offline").Count(&offlineCount)

	fmt.Printf("在线设备: %d\n", onlineCount)
	fmt.Printf("离线设备: %d\n", offlineCount)

	// 心跳统计
	var heartbeatCount int64
	db.Model(&model.DeviceActivityLog{}).Where("action = ?", "heartbeat").Count(&heartbeatCount)
	fmt.Printf("总心跳记录数: %d\n", heartbeatCount)

	// 最近24小时心跳统计
	var recentHeartbeatCount int64
	db.Model(&model.DeviceActivityLog{}).
		Where("action = ? AND created_at > ?", "heartbeat", now.Add(-24*time.Hour)).
		Count(&recentHeartbeatCount)
	fmt.Printf("最近24小时心跳记录数: %d\n", recentHeartbeatCount)
}

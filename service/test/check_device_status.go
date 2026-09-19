package main

import (
	"fmt"
	"log"
	"rvcs/internal/database"
	"rvcs/internal/model"
)

func main() {
	// 初始化数据库连接
	db, err := database.InitSQLite(nil)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 查询所有设备及其状态
	var devices []model.Device
	result := db.Find(&devices)
	if result.Error != nil {
		log.Fatalf("Failed to query devices: %v", result.Error)
	}

	fmt.Printf("Found %d devices:\n", len(devices))
	fmt.Println("ID\tName\t\t\tStatus\t\tLastSeen\t\t\tCreatedAt")
	fmt.Println("-------------------------------------------------------------------------------")
	
	for _, device := range devices {
		lastSeen := "Never"
		if device.LastSeen != nil {
			lastSeen = device.LastSeen.Format("2006-01-02 15:04:05")
		}
		
		createdAt := device.CreatedAt.Format("2006-01-02 15:04:05")
		
		// 截取name显示
		displayName := device.Name
		if len(displayName) > 20 {
			displayName = displayName[:20] + "..."
		}
		
		fmt.Printf("%s\t%-20s\t%s\t%s\t%s\n", 
			device.ID[:8], displayName, device.Status, lastSeen, createdAt)
	}

	// 检查最近的心跳日志
	fmt.Println("\nRecent heartbeat logs:")
	var logs []model.DeviceActivityLog
	result = db.Where("action = ?", "heartbeat").
		Order("created_at DESC").
		Limit(10).
		Find(&logs)
	
	if result.Error != nil {
		log.Printf("Failed to query heartbeat logs: %v", result.Error)
	} else {
		if len(logs) == 0 {
			fmt.Println("No heartbeat logs found")
		} else {
			for _, log := range logs {
				timestamp := log.CreatedAt.Format("2006-01-02 15:04:05")
				fmt.Printf("[%s] Device %s sent heartbeat\n", timestamp, log.DeviceID[:8])
			}
		}
	}
}
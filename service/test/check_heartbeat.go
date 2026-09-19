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

	fmt.Println("=== 设备状态检查 ===")
	
	// 查询所有设备
	var devices []model.Device
	result := db.Find(&devices)
	if result.Error != nil {
		log.Fatalf("Failed to query devices: %v", result.Error)
	}

	fmt.Printf("找到 %d 个设备:\n", len(devices))
	fmt.Println("设备ID\t\t设备名称\t\t\t状态\t\t最后在线时间")
	fmt.Println("------------------------------------------------------------------------")
	
	for _, device := range devices {
		lastSeen := "从未在线"
		if device.LastSeen != nil {
			lastSeen = device.LastSeen.Format("2006-01-02 15:04:05")
		}
		
		// 截取name显示
		displayName := device.Name
		if len(displayName) > 20 {
			displayName = displayName[:20] + "..."
		}
		
		fmt.Printf("%s\t%-20s\t%s\t%s\n", 
			device.ID[:8], displayName, device.Status, lastSeen)
	}

	fmt.Println("\n=== 心跳日志检查 ===")
	
	// 检查最近的心跳日志
	var logs []model.DeviceActivityLog
	result = db.Where("action = ?", "heartbeat").
		Order("created_at DESC").
		Limit(20).
		Find(&logs)
	
	if result.Error != nil {
		log.Printf("查询心跳日志失败: %v", result.Error)
	} else {
		if len(logs) == 0 {
			fmt.Println("未找到心跳日志记录")
		} else {
			fmt.Printf("最近 %d 条心跳记录:\n", len(logs))
			fmt.Println("时间\t\t\t\t设备ID\t\t详情")
			fmt.Println("--------------------------------------------------------------------------------")
			for _, log := range logs {
				timestamp := log.CreatedAt.Format("2006-01-02 15:04:05")
				fmt.Printf("%s\t%s\t%s\n", timestamp, log.DeviceID[:8], log.Details)
			}
		}
	}

	fmt.Println("\n=== 设备活动日志统计 ===")
	
	// 统计各类活动日志
	var stats []struct {
		Action string
		Count  int64
	}
	db.Model(&model.DeviceActivityLog{}).
		Select("action, count(*) as count").
		Group("action").
		Order("count desc").
		Scan(&stats)
	
	fmt.Println("活动类型\t\t数量")
	fmt.Println("------------------------")
	for _, stat := range stats {
		fmt.Printf("%-15s\t%d\n", stat.Action, stat.Count)
	}
}
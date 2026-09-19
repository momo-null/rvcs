package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func main() {
	// 获取Android设备上应用的SharedPreferences内容
	cmd := exec.Command("adb", "shell", "dumpsys", "activity", "service", "com.rvcs.android/.ForegroundCameraService")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get service info: %v", err)
	}

	fmt.Println("Service info:")
	fmt.Println(string(output))

	// 检查SharedPreferences
	fmt.Println("\nChecking SharedPreferences:")
	cmd = exec.Command("adb", "shell", "run-as", "com.rvcs.android", "cat", "/data/data/com.rvcs.android/shared_prefs/device_prefs.xml")
	output, err = cmd.Output()
	if err != nil {
		log.Printf("Failed to read SharedPreferences: %v", err)
		fmt.Println("Trying alternative method...")

		// 替代方法：通过logcat查看注册过程
		cmd = exec.Command("adb", "logcat", "-d", "|", "grep", "-E", "(Registration|SharedPreferences|device_id|access_token)")
		output, err = cmd.Output()
		if err != nil {
			log.Printf("Alternative method failed: %v", err)
		}
	}

	fmt.Println("SharedPreferences content:")
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "device_id") ||
			strings.Contains(line, "access_token") ||
			strings.Contains(line, "is_registered") {
			fmt.Println(line)
		}
	}
}

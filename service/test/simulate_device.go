package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// 模拟设备注册和心跳测试
func main() {
	// 创建不验证SSL证书的HTTP客户端
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	baseURL := "https://localhost:28443"
	
	fmt.Println("=== 设备注册和心跳测试 ===")
	
	// 1. 先生成注册码
	fmt.Println("\n1. 生成注册码...")
	registrationCode := generateRegistrationCode(client, baseURL)
	if registrationCode == "" {
		log.Fatal("无法生成注册码")
	}
	fmt.Printf("生成的注册码: %s\n", registrationCode)
	
	// 2. 使用注册码注册设备
	fmt.Println("\n2. 注册设备...")
	deviceToken, deviceID := registerDevice(client, baseURL, registrationCode)
	if deviceToken == "" {
		log.Fatal("设备注册失败")
	}
	fmt.Printf("设备注册成功!\nDevice ID: %s\nToken: %s\n", deviceID, deviceToken)
	
	// 3. 发送心跳
	fmt.Println("\n3. 发送心跳...")
	sendHeartbeat(client, baseURL, deviceToken, deviceID)
	
	// 4. 持续发送心跳（模拟设备在线）
	fmt.Println("\n4. 持续发送心跳 (按 Ctrl+C 停止)...")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	heartbeatCount := 0
	for {
		select {
		case <-ticker.C:
			heartbeatCount++
			fmt.Printf("发送第 %d 次心跳...\n", heartbeatCount)
			sendHeartbeat(client, baseURL, deviceToken, deviceID)
		}
	}
}

func generateRegistrationCode(client *http.Client, baseURL string) string {
	// 先登录获取管理员token
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	
	loginBody, _ := json.Marshal(loginData)
	loginReq, _ := http.NewRequest("POST", baseURL+"/api/v1/web/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	
	loginResp, err := client.Do(loginReq)
	if err != nil {
		log.Printf("登录失败: %v", err)
		return ""
	}
	defer loginResp.Body.Close()
	
	if loginResp.StatusCode != 200 {
		body, _ := io.ReadAll(loginResp.Body)
		log.Printf("登录失败，状态码: %d, 响应: %s", loginResp.StatusCode, string(body))
		return ""
	}
	
	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	
	token, ok := loginResult["data"].(map[string]interface{})["access_token"].(string)
	if !ok {
		log.Println("无法获取访问令牌")
		return ""
	}
	
	// 生成注册码
	codeData := map[string]int{
		"validity_days": 7,
	}
	
	codeBody, _ := json.Marshal(codeData)
	codeReq, _ := http.NewRequest("POST", baseURL+"/api/v1/web/registration-codes", bytes.NewBuffer(codeBody))
	codeReq.Header.Set("Content-Type", "application/json")
	codeReq.Header.Set("Authorization", "Bearer "+token)
	
	codeResp, err := client.Do(codeReq)
	if err != nil {
		log.Printf("生成注册码失败: %v", err)
		return ""
	}
	defer codeResp.Body.Close()
	
	if codeResp.StatusCode != 200 {
		body, _ := io.ReadAll(codeResp.Body)
		log.Printf("生成注册码失败，状态码: %d, 响应: %s", codeResp.StatusCode, string(body))
		return ""
	}
	
	var codeResult map[string]interface{}
	json.NewDecoder(codeResp.Body).Decode(&codeResult)
	
	code, ok := codeResult["data"].(map[string]interface{})["code"].(string)
	if !ok {
		log.Println("无法获取注册码")
		return ""
	}
	
	return code
}

func registerDevice(client *http.Client, baseURL, registrationCode string) (string, string) {
	// 设备注册数据
	registerData := map[string]interface{}{
		"registration_code": registrationCode,
		"device_name":      "Test_Android_Device",
		"manufacturer":     "Android Emulator",
		"android_version":  "12.0",
		"app_version":      "1.0.0",
	}
	
	body, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/device/register-with-code", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("设备注册请求失败: %v", err)
		return "", ""
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("注册响应状态: %d\n响应内容: %s\n", resp.StatusCode, string(bodyBytes))
	
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", ""
	}
	
	var result map[string]interface{}
	json.Unmarshal(bodyBytes, &result)
	
	// 检查是否有data字段
	data, hasData := result["data"].(map[string]interface{})
	if !hasData {
		// 如果没有data字段，直接使用根级别的字段
		data = result
	}
	
	// 尝试获取各种可能的token字段
	deviceToken := ""
	if token, ok := data["device_token"].(string); ok {
		deviceToken = token
	} else if token, ok := data["access_token"].(string); ok {
		deviceToken = token
	} else if token, ok := data["token"].(string); ok {
		deviceToken = token
	}
	
	if deviceToken == "" {
		log.Println("无法获取设备令牌")
		return "", ""
	}
	
	// 尝试获取设备ID
	deviceID := ""
	if id, ok := data["device_id"].(string); ok {
		deviceID = id
	} else if id, ok := data["id"].(string); ok {
		deviceID = id
	}
	
	if deviceID == "" {
		log.Println("无法获取设备ID")
		return "", ""
	}
	
	return deviceToken, deviceID
}

func sendHeartbeat(client *http.Client, baseURL, token, deviceID string) {
	heartbeatData := map[string]string{
		"status": "online",
	}
	
	body, _ := json.Marshal(heartbeatData)
	req, _ := http.NewRequest("POST", baseURL+"/api/v1/device/heartbeat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("心跳请求失败: %v", err)
		return
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("心跳响应状态: %d\n响应内容: %s\n", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
	
	if resp.StatusCode == 200 {
		fmt.Println("✅ 心跳发送成功!")
	} else {
		fmt.Println("❌ 心跳发送失败!")
	}
}
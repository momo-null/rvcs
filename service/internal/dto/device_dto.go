package dto

import "rvcs/internal/dto/response"

// DeviceRegisterRequest 设备注册请求
type DeviceRegisterRequest struct {
	Name          string `json:"name" binding:"required"`
	IPAddress     string `json:"ip_address,omitempty"`
	Manufacturer  string `json:"manufacturer,omitempty"`
	AndroidVersion string `json:"android_version,omitempty"`
	AppVersion    string `json:"app_version,omitempty"`
}

// DeviceHeartbeatRequest 设备心跳请求
type DeviceHeartbeatRequest struct {
	Status    string `json:"status" binding:"required,oneof=online offline error"`
	Timestamp string `json:"timestamp,omitempty"`
	Health    *DeviceHealthInfo `json:"health,omitempty"`
}

type DeviceHealthInfo struct {
	Battery     int  `json:"battery,omitempty"`
	MemoryUsage int  `json:"memory_usage,omitempty"`
	Streaming   bool `json:"streaming,omitempty"`
}

// DeviceUpdateRequest 设备更新请求
type DeviceUpdateRequest struct {
	Name *string `json:"name,omitempty"`
}

// DeviceControlRequest 设备控制请求
type DeviceControlRequest struct {
	Command    string                 `json:"command" binding:"required"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

type DeviceCommandAckRequest struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
}

// DeviceListResponse 设备列表响应
type DeviceListResponse = response.DeviceListResponse

// DeviceInfo 设备信息
type DeviceInfo = response.DeviceInfo

// ActivityLogListResponse 活动日志列表响应
type ActivityLogListResponse = response.ActivityLogListResponse

// ActivityLogInfo 活动日志信息
type ActivityLogInfo = response.ActivityLogInfo

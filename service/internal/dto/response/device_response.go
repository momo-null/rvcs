package response

import "time"

// DeviceListResponse 设备列表响应
type DeviceListResponse struct {
	Total int          `json:"total"`
	Items []DeviceInfo `json:"items"`
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	IPAddress      string    `json:"ip_address"`
	Manufacturer   string    `json:"manufacturer"`
	AndroidVersion string    `json:"android_version"`
	AppVersion     string    `json:"app_version"`
	BatteryLevel   int       `json:"battery_level"`
	Status         string    `json:"status"`
	LastSeen       time.Time `json:"last_seen"`
	CreatedAt      time.Time `json:"created_at"`
}

// ActivityLogListResponse 活动日志列表响应
type ActivityLogListResponse struct {
	Total int               `json:"total"`
	Items []ActivityLogInfo `json:"items"`
}

// ActivityLogInfo 活动日志信息
type ActivityLogInfo struct {
	ID         string    `json:"id"`
	DeviceID   string    `json:"device_id"`
	DeviceName string    `json:"device_name"`
	Action     string    `json:"action"`
	Details    string    `json:"details,omitempty"`
	User       string    `json:"user,omitempty"`
	RemoteIP   string    `json:"remote_ip,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

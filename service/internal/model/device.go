package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// DeviceStatus 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusOffline DeviceStatus = "offline"
	DeviceStatusError   DeviceStatus = "error"
)

// Device 设备模型
type Device struct {
	ID             string          `json:"id" gorm:"primaryKey;size:64"`
	UserID         *string         `json:"user_id,omitempty" gorm:"size:64;index"`
	Name           string          `json:"name" gorm:"not null;size:100"`
	Status         DeviceStatus    `json:"status" gorm:"default:'offline';index"`
	LastSeen       *time.Time      `json:"last_seen,omitempty"`
	IPAddress      *string         `json:"ip_address,omitempty" gorm:"size:45"`
	Manufacturer   *string         `json:"manufacturer,omitempty" gorm:"size:100"`
	Model          *string         `json:"model,omitempty" gorm:"size:100"`
	AndroidVersion *string         `json:"android_version,omitempty" gorm:"size:20"`
	AppVersion     *string         `json:"app_version,omitempty" gorm:"size:20"`
	BatteryLevel   *int            `json:"battery_level,omitempty" gorm:"index"`
	Capabilities   JSON            `json:"capabilities,omitempty" gorm:"type:json"`
	Settings       JSON            `json:"settings,omitempty" gorm:"type:json"`
	Token          *string         `json:"-" gorm:"size:64;index"` // 设备认证令牌
	PublicKey      *string         `json:"-" gorm:"type:text"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// BeforeCreate 创建前钩子
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// JSON 自定义JSON类型
type JSON json.RawMessage

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	*j = bytes
	return nil
}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// DeviceSession 设备会话
type DeviceSession struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	DeviceID     string     `json:"device_id" gorm:"size:64;not null;index"`
	AccessToken  string     `json:"-" gorm:"type:text;not null"`
	RefreshToken string     `json:"-" gorm:"type:text;not null"`
	IPAddress    *string    `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent    *string    `json:"user_agent,omitempty" gorm:"type:text"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null;index"`
	Revoked      bool       `json:"revoked" gorm:"default:false"`
	CreatedAt    time.Time  `json:"created_at"`
}

// DeviceActivityLog 设备活动日志
type DeviceActivityLog struct {
	ID         uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	DeviceID   string     `json:"device_id" gorm:"size:64;not null;index"`
	Action     string     `json:"action" gorm:"not null;size:50;index"`
	Status     string     `json:"status" gorm:"not null;size:20"`
	Details    JSON       `json:"details,omitempty" gorm:"type:json"`
	RemoteIP   *string    `json:"remote_ip,omitempty" gorm:"size:45"`
	DurationMs *int       `json:"duration_ms,omitempty"`
	CreatedAt  time.Time  `json:"created_at" gorm:"index"`
}

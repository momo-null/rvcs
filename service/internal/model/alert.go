package model

import (
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "info"
	AlertLevelWarning  AlertLevel = "warning"
	AlertLevelError    AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

// Alert 告警模型
type Alert struct {
	ID               string     `json:"id" gorm:"primaryKey;size:64"`
	DeviceID         string     `json:"device_id" gorm:"size:64;not null;index"`
	RuleID           *string    `json:"rule_id,omitempty" gorm:"size:64"`
	EventType        string     `json:"event_type" gorm:"not null;size:50"`
	Level            AlertLevel `json:"level" gorm:"default:'warning';index"`
	Message          *string    `json:"message,omitempty" gorm:"type:text"`
	Data             JSON       `json:"data,omitempty" gorm:"type:json"`
	Acknowledged     bool       `json:"acknowledged" gorm:"default:false;index"`
	AcknowledgedBy   *string    `json:"acknowledged_by,omitempty" gorm:"size:64"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at" gorm:"index"`
}

// AlertRule 告警规则模型
type AlertRule struct {
	ID                   string     `json:"id" gorm:"primaryKey;size:64"`
	UserID               string     `json:"user_id" gorm:"size:64;not null;index"`
	Name                 string     `json:"name" gorm:"not null;size:100"`
	EventType            string     `json:"event_type" gorm:"not null;size:50;index"`
	Level                AlertLevel `json:"level" gorm:"default:'warning'"`
	Conditions           JSON       `json:"conditions" gorm:"type:json;not null"`
	NotificationChannels JSON       `json:"notification_channels,omitempty" gorm:"type:json"`
	Enabled              bool       `json:"enabled" gorm:"default:true"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

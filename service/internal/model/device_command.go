package model

import "time"

type DeviceCommandStatus string

const (
	DeviceCommandStatusPending DeviceCommandStatus = "pending"
	DeviceCommandStatusAcked   DeviceCommandStatus = "acked"
)

// DeviceCommand stores remote control commands for pull-based fallback.
type DeviceCommand struct {
	ID        uint                `json:"id" gorm:"primaryKey;autoIncrement"`
	DeviceID  string              `json:"device_id" gorm:"size:64;not null;index"`
	RequestID string              `json:"request_id" gorm:"size:64;index"`
	Command   string              `json:"command" gorm:"size:64;not null"`
	Payload   JSON                `json:"payload,omitempty" gorm:"type:json"`
	Status    DeviceCommandStatus `json:"status" gorm:"size:20;not null;default:'pending';index"`
	Result    *string             `json:"result,omitempty" gorm:"type:text"`
	CreatedAt time.Time           `json:"created_at" gorm:"index"`
	AckedAt   *time.Time          `json:"acked_at,omitempty" gorm:"index"`
}

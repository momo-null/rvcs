package model

import (
	"time"
)

// DeviceGroup 设备组模型
type DeviceGroup struct {
	ID          string    `json:"id" gorm:"primaryKey;size:64"`
	UserID      string    `json:"user_id" gorm:"size:64;not null;index"`
	Name        string    `json:"name" gorm:"not null;size:100"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Devices     []Device  `json:"devices,omitempty" gorm:"many2many:device_group_members;"`
}

// DeviceGroupMember 设备组成员
type DeviceGroupMember struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	GroupID   string    `json:"group_id" gorm:"size:64;not null;index:idx_group_device"`
	DeviceID  string    `json:"device_id" gorm:"size:64;not null;index:idx_group_device,uniqueIndex"`
	AddedAt   time.Time `json:"added_at"`
}

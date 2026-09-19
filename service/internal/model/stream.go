package model

import (
	"time"
)

// StreamType 视频流类型
type StreamType string

const (
	StreamTypeWebRTC StreamType = "webrtc"
)

// StreamQuality 流质量
type StreamQuality string

const (
	StreamQualityLow    StreamQuality = "low"
	StreamQualityMedium StreamQuality = "medium"
	StreamQualityHigh   StreamQuality = "high"
)

// StreamStatus 流状态
type StreamStatus string

const (
	StreamStatusActive   StreamStatus = "active"
	StreamStatusInactive StreamStatus = "inactive"
	StreamStatusError    StreamStatus = "error"
)

// Stream 视频流模型
type Stream struct {
	ID         string        `json:"id" gorm:"primaryKey;size:64"`
	DeviceID   string        `json:"device_id" gorm:"size:64;not null;index"`
	UserID     string        `json:"user_id" gorm:"size:64;not null;index"`
	StreamType StreamType    `json:"stream_type" gorm:"not null"`
	Quality    StreamQuality `json:"quality" gorm:"default:'medium'"`
	StreamURL  *string       `json:"stream_url,omitempty" gorm:"size:255"`
	AuthToken  *string       `json:"-" gorm:"type:text"`
	Status     StreamStatus  `json:"status" gorm:"default:'active';index"`
	StartedAt  *time.Time    `json:"started_at,omitempty"`
	ExpiresAt  *time.Time    `json:"expires_at,omitempty" gorm:"index"`
	CreatedAt  time.Time     `json:"created_at"`
}

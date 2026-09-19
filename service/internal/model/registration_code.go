package model

import (
	"time"
)

// RegistrationCode 注册码模型 (适配MySQL数据库)
type RegistrationCode struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`     // 主键ID
	Code        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"code"` // 注册码内容
	Status      string    `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // 状态: pending, used, expired, revoked
	DeviceID    *string   `gorm:"type:varchar(50);index" json:"device_id"` // 关联的设备ID（使用后）
	CreatedBy   string    `gorm:"type:varchar(50);not null" json:"created_by"` // 创建者（管理员ID）
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`              // 创建时间
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`              // 更新时间
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`        // 过期时间
	UsedAt      *time.Time `gorm:"index" json:"used_at"`                   // 使用时间
	Description *string   `gorm:"type:text" json:"description"`            // 描述信息
}

// RegistrationCodeStatus 注册码状态枚举
const (
	RegistrationCodePending = "pending"  // 待使用
	RegistrationCodeUsed    = "used"     // 已使用
	RegistrationCodeExpired = "expired"  // 已过期
	RegistrationCodeRevoked = "revoked"  // 已撤销
)

// TableName 指定表名
func (RegistrationCode) TableName() string {
	return "registration_codes"
}

// IsValid 检查注册码是否有效
func (rc *RegistrationCode) IsValid() bool {
	if rc.Status != RegistrationCodePending {
		return false
	}
	
	if time.Now().After(rc.ExpiresAt) {
		rc.Status = RegistrationCodeExpired
		return false
	}
	
	return true
}

// Use 标记注册码为已使用
func (rc *RegistrationCode) Use(deviceID string) {
	now := time.Now()
	rc.Status = RegistrationCodeUsed
	rc.DeviceID = &deviceID
	rc.UsedAt = &now
}

// Revoke 撤销注册码
func (rc *RegistrationCode) Revoke() {
	rc.Status = RegistrationCodeRevoked
}

// IsExpired 检查是否过期
func (rc *RegistrationCode) IsExpired() bool {
	return time.Now().After(rc.ExpiresAt)
}
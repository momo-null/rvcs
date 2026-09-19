package model

import (
	"time"
)

// UserRole 用户角色
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

// User 用户模型
type User struct {
	ID           string     `json:"id" gorm:"primaryKey;size:64"`
	Username     string     `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email        string     `json:"email" gorm:"uniqueIndex;not null;size:100"`
	PasswordHash string     `json:"-" gorm:"not null;size:255"`
	Role         UserRole   `json:"role" gorm:"default:'user'"`
	Status       UserStatus `json:"status" gorm:"default:'active'"`
	Avatar       *string    `json:"avatar,omitempty" gorm:"size:255"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

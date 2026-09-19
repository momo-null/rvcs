package dto

import "rvcs/internal/dto/response"

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse = response.LoginResponse

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse = response.RefreshTokenResponse

// UserInfo 用户信息
type UserInfo = response.UserInfo

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Role     string `form:"role" binding:"omitempty,oneof=admin user"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive locked"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin user"`
	Status   string `json:"status" binding:"required,oneof=active inactive locked"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username *string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    *string `json:"email" binding:"omitempty,email"`
	Role     *string `json:"role" binding:"omitempty,oneof=admin user"`
	Status   *string `json:"status" binding:"omitempty,oneof=active inactive locked"`
}

package dto

// GenerateRegistrationCodeRequest 生成注册码请求
type GenerateRegistrationCodeRequest struct {
	Description  string `json:"description,omitempty" binding:"max=200"` // 描述信息
	ValidityDays int    `json:"validity_days" binding:"required,min=1,max=365"` // 有效期（天）
}

// GenerateRegistrationCodeResponse 生成注册码响应
type GenerateRegistrationCodeResponse struct {
	ID          uint   `json:"id"`          // 注册码ID
	Code        string `json:"code"`        // 注册码内容
	Status      string `json:"status"`      // 状态
	CreatedBy   string `json:"created_by"`  // 创建者
	CreatedAt   string `json:"created_at"`  // 创建时间
	ExpiresAt   string `json:"expires_at"`  // 过期时间
	Description string `json:"description,omitempty"` // 描述信息
}

// RegistrationCodeItem 注册码列表项
type RegistrationCodeItem struct {
	ID          uint    `json:"id"`
	Code        string  `json:"code"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   string  `json:"expires_at"`
	UsedAt      *string `json:"used_at,omitempty"`
	DeviceID    *string `json:"device_id,omitempty"`
	Description string  `json:"description,omitempty"`
}

// GetRegistrationCodeResponse 获取注册码详情响应
type GetRegistrationCodeResponse struct {
	ID          uint    `json:"id"`
	Code        string  `json:"code"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	ExpiresAt   string  `json:"expires_at"`
	UsedAt      *string `json:"used_at,omitempty"`
	DeviceID    *string `json:"device_id,omitempty"`
	Description string  `json:"description,omitempty"`
}

// ListRegistrationCodesResponse 查询注册码列表响应
type ListRegistrationCodesResponse struct {
	Codes      []*RegistrationCodeItem `json:"codes"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
}

// EnhancedDeviceRegisterRequest 增强版设备注册请求（支持注册码）
type EnhancedDeviceRegisterRequest struct {
	RegistrationCode string                 `json:"registration_code" binding:"required"` // 注册码
	DeviceInfo       RegistrationDeviceInfo `json:"device_info"`                          // 设备信息
	PublicKey        string                 `json:"public_key,omitempty"`                 // 设备公钥（可选）
	Signature        string                 `json:"signature,omitempty"`                  // 签名（可选）
}

// RegistrationDeviceInfo 注册用设备信息（避免命名冲突）
type RegistrationDeviceInfo struct {
	Name        string `json:"name" binding:"required"`        // 设备名称
	Type        string `json:"type" binding:"required"`        // 设备类型
	Model       string `json:"model,omitempty"`                // 设备型号
	Location    string `json:"location,omitempty"`             // 位置信息
	Description string `json:"description,omitempty"`          // 描述信息
	OSVersion   string `json:"os_version,omitempty"`           // 操作系统版本
	AppVersion  string `json:"app_version,omitempty"`          // 应用版本
}

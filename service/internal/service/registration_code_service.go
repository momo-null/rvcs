package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"rvcs/internal/core"
	"rvcs/internal/model"
	"rvcs/internal/repository"
)

// RegistrationCodeService 注册码服务接口
type RegistrationCodeService interface {
	GenerateCode(ctx context.Context, createdBy string, description string, validityDays int) (*model.RegistrationCode, error)
	ValidateCode(ctx context.Context, code string) (*model.RegistrationCode, error)
	UseCode(ctx context.Context, code string, deviceID string) error
	ListCodes(ctx context.Context, page, pageSize int, status string) ([]*model.RegistrationCode, int64, error)
	GetCodeByID(ctx context.Context, id string) (*model.RegistrationCode, error)
	DeleteCode(ctx context.Context, id string) error
	RevokeCode(ctx context.Context, id string) error
	ResetCode(ctx context.Context, id string, extendDays *int) error
	CleanupExpiredCodes(ctx context.Context) (int64, error)
}

// registrationCodeService 注册码服务实现
type registrationCodeService struct {
	repo      repository.RegistrationCodeRepository
	generator *core.RegistrationCodeGenerator
	reusable  bool // 注册码是否可重复使用
}

// NewRegistrationCodeService 创建注册码服务
func NewRegistrationCodeService(repo repository.RegistrationCodeRepository, reusable bool) (RegistrationCodeService, error) {
	generator, err := core.NewRegistrationCodeGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create code generator: %v", err)
	}

	return &registrationCodeService{
		repo:      repo,
		generator: generator,
		reusable:  reusable,
	}, nil
}

// GenerateCode 生成注册码
func (s *registrationCodeService) GenerateCode(ctx context.Context, createdBy string, description string, validityDays int) (*model.RegistrationCode, error) {
	// 生成注册码
	codeStr, err := s.generator.GenerateCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate registration code: %v", err)
	}

	// 创建注册码对象
	now := time.Now()
	descriptionPtr := &description
	code := &model.RegistrationCode{
		Code:        codeStr,
		Status:      model.RegistrationCodePending,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   now.Add(time.Duration(validityDays) * 24 * time.Hour),
		Description: descriptionPtr,
	}

	// 保存到数据库
	if err := s.repo.Create(ctx, code); err != nil {
		return nil, fmt.Errorf("failed to save registration code: %v", err)
	}

	return code, nil
}

// ValidateCode 验证注册码有效性
func (s *registrationCodeService) ValidateCode(ctx context.Context, codeStr string) (*model.RegistrationCode, error) {
	// 格式验证
	if !s.generator.ValidateCode(codeStr) {
		return nil, fmt.Errorf("invalid registration code format")
	}

	// 数据库查询
	code, err := s.repo.GetByCode(ctx, codeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query registration code: %v", err)
	}

	if code == nil {
		return nil, fmt.Errorf("registration code not found")
	}

	// 检查过期时间（过期码永远无效）
	if code.IsExpired() {
		return nil, fmt.Errorf("registration code has expired")
	}

	// 状态验证
	// 在可重复使用模式下，允许 used 状态的注册码通过验证
	// 这样设备token过期后可以用同一注册码重新注册，返回同一个设备
	if !s.reusable && !code.IsValid() {
		return nil, fmt.Errorf("registration code is not valid (status: %s)", code.Status)
	} else if !s.reusable && code.Status != model.RegistrationCodePending {
		// 非重复使用模式下，只接受 pending 状态
		return nil, fmt.Errorf("registration code is not valid (status: %s)", code.Status)
	} else if s.reusable && code.Status != model.RegistrationCodePending && code.Status != model.RegistrationCodeUsed {
		// 重复使用模式下，只接受 pending 和 used 状态
		return nil, fmt.Errorf("registration code is not valid (status: %s)", code.Status)
	}

	return code, nil
}

// UseCode 使用注册码
func (s *registrationCodeService) UseCode(ctx context.Context, codeStr string, deviceID string) error {
	// 验证注册码
	code, err := s.ValidateCode(ctx, codeStr)
	if err != nil {
		return err
	}

	// 如果注册码不可重复使用，则标记为已使用
	if !s.reusable {
		code.Use(deviceID)

		// 更新数据库
		updates := map[string]interface{}{
			"status":    code.Status,
			"device_id": code.DeviceID,
			"used_at":   code.UsedAt,
		}

		if err := s.repo.Update(ctx, code.ID, updates); err != nil {
			return fmt.Errorf("failed to update registration code: %v", err)
		}
	} else {
		// 可重复使用模式：只保存 device_id 用于识别已注册设备，但不改变状态
		// 这样同一注册码可以多次使用，但返回同一个设备
		if code.DeviceID == nil {
			// 首次使用，保存 device_id
			updates := map[string]interface{}{
				"device_id": deviceID,
				"used_at":   time.Now(),
			}

			if err := s.repo.Update(ctx, code.ID, updates); err != nil {
				return fmt.Errorf("failed to update registration code: %v", err)
			}
		}
		// 如果 device_id 已存在，不做任何操作（说明是同一个设备重新注册）
	}

	return nil
}

// ListCodes 查询注册码列表
func (s *registrationCodeService) ListCodes(ctx context.Context, page, pageSize int, status string) ([]*model.RegistrationCode, int64, error) {
	filter := make(map[string]interface{})

	if status != "" {
		filter["status"] = status
	}

	return s.repo.List(ctx, filter, page, pageSize)
}

// GetCodeByID 根据ID获取注册码
func (s *registrationCodeService) GetCodeByID(ctx context.Context, id string) (*model.RegistrationCode, error) {
	// 解析ID为uint
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %v", err)
	}

	return s.repo.GetByID(ctx, uint(uintID))
}

// DeleteCode 删除注册码
func (s *registrationCodeService) DeleteCode(ctx context.Context, id string) error {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid ID format: %v", err)
	}

	return s.repo.Delete(ctx, uint(uintID))
}

// RevokeCode 撤销注册码
func (s *registrationCodeService) RevokeCode(ctx context.Context, id string) error {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid ID format: %v", err)
	}

	// 获取注册码
	code, err := s.repo.GetByID(ctx, uint(uintID))
	if err != nil {
		return fmt.Errorf("failed to get registration code: %v", err)
	}

	if code == nil {
		return fmt.Errorf("registration code not found")
	}

	// 检查状态
	if code.Status == model.RegistrationCodeUsed {
		return fmt.Errorf("cannot revoke used registration code")
	}

	// 撤销注册码
	code.Revoke()

	// 更新数据库
	updates := map[string]interface{}{
		"status": code.Status,
	}

	return s.repo.Update(ctx, uint(uintID), updates)
}

// CleanupExpiredCodes 清理过期注册码
func (s *registrationCodeService) CleanupExpiredCodes(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpired(ctx)
}

// ResetCode 重置注册码
func (s *registrationCodeService) ResetCode(ctx context.Context, id string, extendDays *int) error {
	uintID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid ID format: %v", err)
	}

	// 获取注册码
	code, err := s.repo.GetByID(ctx, uint(uintID))
	if err != nil {
		return fmt.Errorf("failed to get registration code: %v", err)
	}

	if code == nil {
		return fmt.Errorf("registration code not found")
	}

	// 过期的注册码不建议重置
	if code.IsExpired() {
		return fmt.Errorf("cannot reset expired registration code, please generate a new one")
	}

	// 重置注册码状态
	code.Status = model.RegistrationCodePending
	code.DeviceID = nil
	code.UsedAt = nil

	// 如果需要延长时间
	if extendDays != nil && *extendDays > 0 {
		// 延长时间: 在当前过期时间基础上增加指定天数
		code.ExpiresAt = code.ExpiresAt.Add(time.Duration(*extendDays) * 24 * time.Hour)
	}

	// 更新数据库
	updates := map[string]interface{}{
		"status":     code.Status,
		"device_id":  code.DeviceID,
		"used_at":    code.UsedAt,
		"expires_at": code.ExpiresAt,
	}

	return s.repo.Update(ctx, uint(uintID), updates)
}

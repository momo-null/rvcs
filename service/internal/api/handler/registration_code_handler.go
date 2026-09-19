package handler

import (
	"strconv"
	"time"

	"rvcs/internal/api/response"
	"rvcs/internal/dto"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RegistrationCodeHandler 注册码处理器
type RegistrationCodeHandler struct {
	registrationCodeService service.RegistrationCodeService
}

// NewRegistrationCodeHandler 创建注册码处理器
func NewRegistrationCodeHandler(regCodeService service.RegistrationCodeService) *RegistrationCodeHandler {
	return &RegistrationCodeHandler{
		registrationCodeService: regCodeService,
	}
}

// GenerateCode 生成注册码
func (h *RegistrationCodeHandler) GenerateCode(c *gin.Context) {
	var req dto.GenerateRegistrationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 从上下文获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// 默认有效期30天（如果未指定）
	if req.ValidityDays == 0 {
		req.ValidityDays = 30
	}

	logger.Info("Generate registration code request",
		zap.String("user_id", userID.(string)),
		zap.Int("validity_days", req.ValidityDays),
		zap.String("description", req.Description),
	)

	// 生成注册码
	code, err := h.registrationCodeService.GenerateCode(c.Request.Context(), userID.(string), req.Description, req.ValidityDays)
	if err != nil {
		logger.Warn("Failed to generate registration code",
			zap.String("user_id", userID.(string)),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	logger.Info("Registration code generated successfully", zap.Uint("code_id", code.ID))

	// 返回响应
	response.Success(c, dto.GenerateRegistrationCodeResponse{
		ID:          code.ID,
		Code:        code.Code,
		Status:      code.Status,
		CreatedBy:   code.CreatedBy,
		CreatedAt:   code.CreatedAt.Format(time.RFC3339),
		ExpiresAt:   code.ExpiresAt.Format(time.RFC3339),
		Description: stringValue(code.Description),
	})
}

// ListCodes 查询注册码列表
func (h *RegistrationCodeHandler) ListCodes(c *gin.Context) {
	page := 1
	pageSize := 10
	status := c.Query("status")

	// 解析查询参数
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	logger.Debug("List registration codes request",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("status", status),
	)

	// 查询注册码列表
	codes, total, err := h.registrationCodeService.ListCodes(c.Request.Context(), page, pageSize, status)
	if err != nil {
		logger.Error("Failed to list registration codes", zap.Error(err))
		response.Error(c, 500, err.Error())
		return
	}

	// 转换为响应格式
	items := make([]*dto.RegistrationCodeItem, len(codes))
	for i, code := range codes {
		items[i] = h.convertToItem(code)
	}

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response.Success(c, dto.ListRegistrationCodesResponse{
		Codes:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}

// GetCode 获取注册码详情
func (h *RegistrationCodeHandler) GetCode(c *gin.Context) {
	id := c.Param("id")

	logger.Debug("Get registration code request", zap.String("code_id", id))

	code, err := h.registrationCodeService.GetCodeByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get registration code",
			zap.String("code_id", id),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	if code == nil {
		logger.Warn("Registration code not found", zap.String("code_id", id))
		response.Error(c, 404, "registration code not found")
		return
	}

	response.Success(c, h.convertToDetailResponse(code))
}

// DeleteCode 删除注册码
func (h *RegistrationCodeHandler) DeleteCode(c *gin.Context) {
	id := c.Param("id")

	// 检查注册码是否存在且可删除
	code, err := h.registrationCodeService.GetCodeByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("Failed to get registration code",
			zap.String("code_id", id),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	if code == nil {
		logger.Warn("Registration code not found", zap.String("code_id", id))
		response.Error(c, 404, "registration code not found")
		return
	}

	// 已使用的注册码不能删除
	if code.Status == model.RegistrationCodeUsed {
		logger.Warn("Attempted to delete used registration code", zap.String("code_id", id))
		response.BadRequest(c, "cannot delete used registration code")
		return
	}

	logger.Info("Delete registration code request", zap.String("code_id", id))

	// 执行删除
	if err := h.registrationCodeService.DeleteCode(c.Request.Context(), id); err != nil {
		logger.Warn("Failed to delete registration code",
			zap.String("code_id", id),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	logger.Info("Registration code deleted successfully", zap.String("code_id", id))
	response.Success(c, gin.H{"message": "registration code deleted successfully"})
}

// RevokeCode 撤销注册码
func (h *RegistrationCodeHandler) RevokeCode(c *gin.Context) {
	id := c.Param("id")

	logger.Info("Revoke registration code request", zap.String("code_id", id))

	if err := h.registrationCodeService.RevokeCode(c.Request.Context(), id); err != nil {
		logger.Warn("Failed to revoke registration code",
			zap.String("code_id", id),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	logger.Info("Registration code revoked successfully", zap.String("code_id", id))
	response.Success(c, gin.H{"message": "registration code revoked successfully"})
}

// ResetCodeRequest 重置注册码请求
type ResetCodeRequest struct {
	ExtendDays *int `json:"extend_days,omitempty"` // 延长天数,可选
}

// ResetCode 重置注册码
func (h *RegistrationCodeHandler) ResetCode(c *gin.Context) {
	id := c.Param("id")

	var req ResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 验证延长天数
	if req.ExtendDays != nil && *req.ExtendDays < 0 {
		response.BadRequest(c, "extend_days must be non-negative")
		return
	}

	logger.Info("Reset registration code request",
		zap.String("code_id", id),
		zap.Any("extend_days", req.ExtendDays),
	)

	if err := h.registrationCodeService.ResetCode(c.Request.Context(), id, req.ExtendDays); err != nil {
		logger.Warn("Failed to reset registration code",
			zap.String("code_id", id),
			zap.String("error", err.Error()),
		)
		response.Error(c, 500, err.Error())
		return
	}

	logger.Info("Registration code reset successfully", zap.String("code_id", id))
	response.Success(c, gin.H{"message": "registration code reset successfully"})
}

// CleanupExpiredCodes 清理过期注册码
func (h *RegistrationCodeHandler) CleanupExpiredCodes(c *gin.Context) {
	logger.Info("Cleanup expired registration codes request")

	count, err := h.registrationCodeService.CleanupExpiredCodes(c.Request.Context())
	if err != nil {
		logger.Error("Failed to cleanup expired registration codes", zap.Error(err))
		response.Error(c, 500, err.Error())
		return
	}

	logger.Info("Expired registration codes cleaned up successfully", zap.Int64("deleted_count", count))
	response.Success(c, gin.H{
		"message":       "expired registration codes cleaned up successfully",
		"deleted_count": count,
	})
}

// convertToItem 转换为列表项
func (h *RegistrationCodeHandler) convertToItem(code *model.RegistrationCode) *dto.RegistrationCodeItem {
	item := &dto.RegistrationCodeItem{
		ID:          code.ID,
		Code:        code.Code,
		Status:      code.Status,
		CreatedBy:   code.CreatedBy,
		CreatedAt:   code.CreatedAt.Format(time.RFC3339),
		ExpiresAt:   code.ExpiresAt.Format(time.RFC3339),
		Description: stringValue(code.Description),
	}

	if code.UsedAt != nil {
		usedAt := code.UsedAt.Format(time.RFC3339)
		item.UsedAt = &usedAt
	}

	if code.DeviceID != nil {
		item.DeviceID = code.DeviceID
	}

	return item
}

// convertToDetailResponse 转换为详情响应
func (h *RegistrationCodeHandler) convertToDetailResponse(code *model.RegistrationCode) *dto.GetRegistrationCodeResponse {
	resp := &dto.GetRegistrationCodeResponse{
		ID:          code.ID,
		Code:        code.Code,
		Status:      code.Status,
		CreatedBy:   code.CreatedBy,
		CreatedAt:   code.CreatedAt.Format(time.RFC3339),
		ExpiresAt:   code.ExpiresAt.Format(time.RFC3339),
		Description: stringValue(code.Description),
	}

	if code.UsedAt != nil {
		usedAt := code.UsedAt.Format(time.RFC3339)
		resp.UsedAt = &usedAt
	}

	if code.DeviceID != nil {
		resp.DeviceID = code.DeviceID
	}

	return resp
}

// stringValue 安全获取字符串指针的值
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

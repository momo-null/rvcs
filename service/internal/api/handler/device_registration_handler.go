package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"rvcs/internal/api/response"
	"rvcs/internal/logger"
	"rvcs/internal/service"
)

// RegistrationWithCodeRequest 使用注册码注册设备的请求
type RegistrationWithCodeRequest struct {
	RegistrationCode string `json:"registration_code" binding:"required"`
	DeviceName       string `json:"device_name" binding:"required"`
	Manufacturer     string `json:"manufacturer"`
	AndroidVersion   string `json:"android_version"`
	AppVersion       string `json:"app_version"`
}

// RegistrationResponse 设备注册响应
type RegistrationResponse struct {
	DeviceToken  string `json:"device_token"`
	RefreshToken string `json:"refresh_token"`
	DeviceID     string `json:"device_id"`
	ExpiresAt    string `json:"expires_at"`
}

// RegisterWithCode 使用注册码注册设备
// @Summary 使用注册码注册设备
// @Tags Device
// @Accept json
// @Produce json
// @Param request body RegistrationWithCodeRequest true "注册请求"
// @Success 200 {object} Response
// @Router /api/v1/device/register-with-code [post]
func (h *DeviceHandler) RegisterWithCode(c *gin.Context) {
	var req RegistrationWithCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid registration request", zap.Error(err))
		response.DeviceError(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	// 调用服务层进行注册码验证和设备注册
	result, err := h.deviceService.RegisterWithCode(&service.RegistrationWithCodeRequest{
		RegistrationCode: req.RegistrationCode,
		DeviceName:       req.DeviceName,
		Manufacturer:     req.Manufacturer,
		AndroidVersion:   req.AndroidVersion,
		AppVersion:       req.AppVersion,
		ClientIP:         c.ClientIP(),
		UserAgent:        c.Request.UserAgent(),
	})
	if err != nil {
		logger.Error("Device registration failed", zap.Error(err))
		response.DeviceError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 生成设备Token和刷新Token
	deviceToken, err := h.jwtManager.GenerateDeviceToken(result.DeviceID)
	if err != nil {
		logger.Error("Failed to generate device token", zap.Error(err))
		response.DeviceError(c, http.StatusInternalServerError, "Failed to generate device token")
		return
	}

	refreshToken, err := h.jwtManager.GenerateDeviceToken(result.DeviceID) // 使用相同方法生成刷新token
	if err != nil {
		logger.Error("Failed to generate refresh token", zap.Error(err))
		response.DeviceError(c, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	resp := RegistrationResponse{
		DeviceToken:  deviceToken,
		RefreshToken: refreshToken,
		DeviceID:     result.DeviceID,
		ExpiresAt:    time.Now().Add(2 * time.Hour).Format(time.RFC3339),
	}

	logger.Info("Device registered successfully with code",
		zap.String("device_id", result.DeviceID),
		zap.String("registration_code", req.RegistrationCode))

	response.DeviceSuccess(c, resp)
}

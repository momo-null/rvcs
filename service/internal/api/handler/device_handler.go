package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"rvcs/internal/api/response"
	"rvcs/internal/config"
	"rvcs/internal/dto"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/security"
	"rvcs/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
)

type DeviceHandler struct {
	deviceService           service.DeviceService
	registrationCodeService service.RegistrationCodeService
	jwtManager              *security.JWTManager
	liveKitConfig           *config.LiveKitConfig
}

func NewDeviceHandler(
	deviceService service.DeviceService,
	registrationCodeService service.RegistrationCodeService,
	jwtManager *security.JWTManager,
	liveKitConfig *config.LiveKitConfig,
) *DeviceHandler {
	return &DeviceHandler{
		deviceService:           deviceService,
		registrationCodeService: registrationCodeService,
		jwtManager:              jwtManager,
		liveKitConfig:           liveKitConfig,
	}
}

// Register 增强版设备注册（支持注册码）
func (h *DeviceHandler) Register(c *gin.Context) {
	var req dto.EnhancedDeviceRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.DeviceError(c, http.StatusBadRequest, err.Error())
		return
	}

	logger.Info("Device register request",
		zap.String("registration_code", req.RegistrationCode),
		zap.String("device_name", req.DeviceInfo.Name),
		zap.String("device_model", req.DeviceInfo.Model),
		zap.String("client_ip", c.ClientIP()),
	)

	// 验证注册码
	code, err := h.registrationCodeService.ValidateCode(c.Request.Context(), req.RegistrationCode)
	if err != nil {
		logger.Warn("Registration code validation failed",
			zap.String("registration_code", req.RegistrationCode),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusUnauthorized, err.Error())
		return
	}

	logger.Debug("Registration code validation successful", zap.Uint("code_id", code.ID))

	var existingDevice *model.Device

	// 1. 检查注册码是否已被使用（优先级最高）
	if code.DeviceID != nil && *code.DeviceID != "" {
		// 注册码已被使用，获取该设备
		deviceWithCode, err := h.deviceService.GetByID(*code.DeviceID)
		if err != nil {
			response.DeviceError(c, http.StatusInternalServerError, "Failed to get existing device")
			return
		}

		// 验证是否是同一设备（通过name和model判断）
		if req.DeviceInfo.Model != "" {
			if deviceWithCode.Name == req.DeviceInfo.Name && safeString(deviceWithCode.Model) == req.DeviceInfo.Model {
				// 是同一设备，允许重新注册
				existingDevice = deviceWithCode
				logger.Info("Same device re-registering with existing code",
					zap.String("device_id", deviceWithCode.ID),
					zap.String("registration_code", req.RegistrationCode),
				)
			} else {
				// 不同设备尝试使用已用的注册码，拒绝
				logger.Warn("Different device trying to use used registration code",
					zap.String("device_name", req.DeviceInfo.Name),
					zap.String("device_model", req.DeviceInfo.Model),
					zap.String("registration_code", req.RegistrationCode),
					zap.String("existing_device_id", deviceWithCode.ID),
					zap.String("existing_device_name", deviceWithCode.Name),
					zap.String("existing_device_model", safeString(deviceWithCode.Model)),
				)
				response.DeviceError(c, http.StatusConflict,
					"Registration code already used by another device")
				return
			}
		} else {
			// 没有提供model信息，保守策略：拒绝
			logger.Warn("Device with no model info trying to use used registration code",
				zap.String("registration_code", req.RegistrationCode),
			)
			response.DeviceError(c, http.StatusConflict,
				"Registration code already used and device info insufficient for verification")
			return
		}
	} else if req.DeviceInfo.Model != "" {
		// 2. 尝试根据设备名称和型号查找（用于新注册码识别同一设备）
		existingDevice, err = h.deviceService.FindByNameAndModel(req.DeviceInfo.Name, req.DeviceInfo.Model)
		if err == nil {
			logger.Info("Found existing device by name and model",
				zap.String("device_id", existingDevice.ID),
				zap.String("registration_code", req.RegistrationCode),
			)
		}
	}

	// 如果找到已存在的设备，返回它和新token
	if existingDevice != nil {
		// 为已存在的设备生成新token
		deviceToken, refreshToken, err := h.jwtManager.GenerateDeviceTokenPair(existingDevice.ID)
		if err != nil {
			logger.Error("Failed to generate device tokens",
				zap.String("device_id", existingDevice.ID),
				zap.String("error", err.Error()),
			)
			response.DeviceError(c, http.StatusInternalServerError, "Failed to generate device tokens")
			return
		}

		// 标记注册码与该设备关联
		if err := h.registrationCodeService.UseCode(c.Request.Context(), req.RegistrationCode, existingDevice.ID); err != nil {
			logger.Error("Failed to link registration code to device",
				zap.String("registration_code", req.RegistrationCode),
				zap.String("device_id", existingDevice.ID),
				zap.String("error", err.Error()),
			)
		}

		logger.Info("Returning existing device for registration code",
			zap.String("device_id", existingDevice.ID),
			zap.String("registration_code", req.RegistrationCode),
		)

		// 返回已存在的设备信息（包含新token）
		resp := gin.H{
			"id":                 existingDevice.ID,
			"name":               existingDevice.Name,
			"manufacturer":       existingDevice.Manufacturer,
			"android_version":    existingDevice.AndroidVersion,
			"app_version":        existingDevice.AppVersion,
			"status":             existingDevice.Status,
			"created_at":         existingDevice.CreatedAt,
			"device_token":       deviceToken,
			"refresh_token":      refreshToken,
			"already_registered": true,
		}

		response.DeviceCreated(c, resp)
		return
	}

	// 构建标准设备注册请求
	deviceReq := dto.DeviceRegisterRequest{
		Name:           req.DeviceInfo.Name,
		Manufacturer:   req.DeviceInfo.Model,
		AndroidVersion: req.DeviceInfo.OSVersion,
		AppVersion:     req.DeviceInfo.AppVersion,
	}

	// 调用服务层注册设备
	device, err := h.deviceService.Register(&deviceReq)
	if err != nil {
		logger.Warn("Failed to register device",
			zap.String("device_name", req.DeviceInfo.Name),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusBadRequest, err.Error())
		return
	}

	// 生成设备Token和刷新Token
	deviceToken, refreshToken, err := h.jwtManager.GenerateDeviceTokenPair(device.ID)
	if err != nil {
		logger.Error("Failed to generate device tokens",
			zap.String("device_id", device.ID),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusInternalServerError, "Failed to generate device tokens")
		return
	}

	// 标记注册码为已使用
	if err := h.registrationCodeService.UseCode(c.Request.Context(), req.RegistrationCode, device.ID); err != nil {
		// 记录错误但不中断注册流程
		logger.Warn("Failed to mark registration code as used",
			zap.String("registration_code", req.RegistrationCode),
			zap.String("device_id", device.ID),
			zap.String("error", err.Error()),
		)
	}

	logger.Info("Device registered successfully", zap.String("device_id", device.ID))

	// 返回注册响应（包含token）
	resp := gin.H{
		"id":                 device.ID,
		"name":               device.Name,
		"manufacturer":       device.Manufacturer,
		"android_version":    device.AndroidVersion,
		"app_version":        device.AppVersion,
		"status":             device.Status,
		"created_at":         device.CreatedAt,
		"device_token":       deviceToken,
		"refresh_token":      refreshToken,
		"already_registered": false,
	}

	response.DeviceCreated(c, resp)
}

// enhancedRegister 增强版注册（支持RSA安全认证）
func (h *DeviceHandler) enhancedRegister(c *gin.Context, device *model.Device, req dto.EnhancedDeviceRegisterRequest) {
	// 简化处理，跳过RSA验证
	response.Success(c, gin.H{
		"device":        device,
		"security_mode": "ENHANCED",
	})
}

// CreateDevice 管理员创建设备
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req dto.DeviceRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// 调用服务层创建设备
	device, err := h.deviceService.Create(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, device)
}

// Heartbeat 设备心跳
func (h *DeviceHandler) Heartbeat(c *gin.Context) {
	var req dto.DeviceHeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid heartbeat request format",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		response.DeviceError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文获取设备ID
	deviceID, exists := c.Get("device_id")
	if !exists {
		logger.Warn("Heartbeat from unauthenticated device",
			zap.String("client_ip", c.ClientIP()),
		)
		response.Unauthorized(c, "Device not authenticated")
		return
	}

	deviceIDStr := deviceID.(string)

	logger.Debug("Device heartbeat received",
		zap.String("device_id", deviceIDStr),
		zap.String("status", req.Status),
	)

	// 调用服务层处理心跳
	err := h.deviceService.Heartbeat(deviceIDStr, &req, c.ClientIP())
	if err != nil {
		logger.Error("Failed to process device heartbeat",
			zap.String("device_id", deviceIDStr),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Debug("Device heartbeat processed successfully", zap.String("device_id", deviceIDStr))
	response.DeviceSuccess(c, gin.H{"status": "ok"})
}

// PullPendingCommands 设备拉取待执行命令
// GET /api/v1/device/commands/pending?limit=20
func (h *DeviceHandler) PullPendingCommands(c *gin.Context) {
	deviceID, exists := c.Get("device_id")
	if !exists {
		response.DeviceError(c, http.StatusUnauthorized, "Device not authenticated")
		return
	}
	deviceIDStr := deviceID.(string)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.deviceService.GetPendingCommands(deviceIDStr, limit)
	if err != nil {
		response.DeviceError(c, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]gin.H, 0, len(items))
	for _, item := range items {
		payload := map[string]interface{}{}
		if len(item.Payload) > 0 {
			_ = json.Unmarshal(item.Payload, &payload)
		}
		resp = append(resp, gin.H{
			"id":         item.ID,
			"request_id": item.RequestID,
			"command":    item.Command,
			"payload":    payload,
			"created_at": item.CreatedAt,
		})
	}

	response.DeviceSuccess(c, gin.H{
		"items": resp,
	})
}

// AckCommand 设备回执命令执行结果
// POST /api/v1/device/commands/:id/ack
func (h *DeviceHandler) AckCommand(c *gin.Context) {
	deviceID, exists := c.Get("device_id")
	if !exists {
		response.DeviceError(c, http.StatusUnauthorized, "Device not authenticated")
		return
	}
	deviceIDStr := deviceID.(string)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.DeviceError(c, http.StatusBadRequest, "invalid command id")
		return
	}

	var req dto.DeviceCommandAckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.DeviceError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	if err := h.deviceService.AckCommand(deviceIDStr, uint(id), req.Success, req.Result); err != nil {
		response.DeviceError(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.DeviceSuccess(c, gin.H{"status": "ok"})
}

// GetDevice 获取设备详情
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id := c.Param("id")

	logger.Debug("Get device request", zap.String("device_id", id))

	// 调用服务层获取设备
	device, err := h.deviceService.GetByID(id)
	if err != nil {
		logger.Warn("Device not found", zap.String("device_id", id))
		response.NotFound(c, "Device not found")
		return
	}

	response.Success(c, device)
}

// ListDevices 获取设备列表
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	logger.Debug("Get device list request",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("status", status),
	)

	// 调用服务层获取设备列表
	resp, err := h.deviceService.List(page, pageSize, status)
	if err != nil {
		logger.Error("Failed to get device list", zap.Error(err))
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// safeString 安全地获取string指针的值
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UpdateDevice 更新设备信息
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id := c.Param("id")
	var req dto.DeviceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("Update device request", zap.String("device_id", id))

	// 调用服务层更新设备
	device, err := h.deviceService.Update(id, &req)
	if err != nil {
		logger.Warn("Failed to update device",
			zap.String("device_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("Device updated successfully", zap.String("device_id", id))
	response.Success(c, device)
}

// DeleteDevice 删除设备
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id := c.Param("id")

	logger.Info("Delete device request", zap.String("device_id", id))

	// 调用服务层删除设备
	err := h.deviceService.Delete(id)
	if err != nil {
		logger.Warn("Failed to delete device",
			zap.String("device_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("Device deleted successfully", zap.String("device_id", id))
	response.Success(c, gin.H{"message": "Device deleted successfully"})
}

// RefreshDeviceToken 刷新设备Token
func (h *DeviceHandler) RefreshDeviceToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.DeviceError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	logger.Debug("Device token refresh request", zap.String("client_ip", c.ClientIP()))

	// 验证refresh token并获取claims
	claims, err := h.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		logger.Warn("Device token refresh failed",
			zap.String("error", err.Error()),
			zap.Bool("is_expired", err == security.ErrExpiredToken),
		)
		if err == security.ErrExpiredToken {
			response.DeviceError(c, http.StatusUnauthorized, "Refresh token expired")
		} else {
			response.DeviceError(c, http.StatusUnauthorized, "Invalid refresh token")
		}
		return
	}

	// 确保是设备token
	if claims.DeviceID == "" {
		logger.Warn("Attempted to refresh non-device token")
		response.DeviceError(c, http.StatusUnauthorized, "Not a device token")
		return
	}

	// 生成新的token对
	newAccessToken, newRefreshToken, err := h.jwtManager.GenerateDeviceTokenPair(claims.DeviceID)
	if err != nil {
		logger.Error("Failed to generate new device tokens",
			zap.String("device_id", claims.DeviceID),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusInternalServerError, "Failed to generate new tokens")
		return
	}

	logger.Info("Device token refreshed successfully", zap.String("device_id", claims.DeviceID))
	response.DeviceSuccess(c, gin.H{
		"device_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// ControlDevice 控制设备
func (h *DeviceHandler) ControlDevice(c *gin.Context) {
	id := c.Param("id")
	var req dto.DeviceControlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("Control device request",
		zap.String("device_id", id),
		zap.String("command", req.Command),
	)

	// 调用服务层控制设备
	err := h.deviceService.Control(id, &req)
	if err != nil {
		logger.Warn("Failed to control device",
			zap.String("device_id", id),
			zap.String("command", req.Command),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("Device control command sent successfully",
		zap.String("device_id", id),
		zap.String("command", req.Command),
	)
	response.Success(c, gin.H{"message": "Command sent successfully"})
}

// GetDeviceLogs 获取设备日志
func (h *DeviceHandler) GetDeviceLogs(c *gin.Context) {
	id := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logger.Debug("Get device logs request",
		zap.String("device_id", id),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 调用服务层获取设备日志
	resp, err := h.deviceService.GetLogs(id, page, pageSize)
	if err != nil {
		logger.Error("Failed to get device logs",
			zap.String("device_id", id),
			zap.String("error", err.Error()),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

func (h *DeviceHandler) GetAllLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logger.Debug("Get all device logs request",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
	)

	// 调用服务层获取所有设备日志
	resp, err := h.deviceService.GetAllLogs(page, pageSize)
	if err != nil {
		logger.Error("Failed to get all device logs", zap.Error(err))
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, resp)
}

// GetLiveKitToken 生成LiveKit连接Token
// GET /api/v1/web/devices/:id/livekit-token
func (h *DeviceHandler) GetLiveKitToken(c *gin.Context) {
	deviceID := c.Param("id")
	role := c.DefaultQuery("role", "viewer")

	if role != "viewer" && role != "talkback" {
		response.BadRequest(c, "invalid role")
		return
	}

	logger.Info("Get LiveKit token request",
		zap.String("device_id", deviceID),
		zap.String("role", role),
		zap.String("client_ip", c.ClientIP()),
	)

	// 检查设备是否存在
	_, err := h.deviceService.GetByID(deviceID)
	if err != nil {
		logger.Warn("Device not found",
			zap.String("device_id", deviceID),
		)
		response.NotFound(c, "Device not found")
		return
	}

	// 生成LiveKit Token（Web客户端用）
	token, err := h.generateLiveKitToken(deviceID, role)
	if err != nil {
		logger.Error("Failed to generate LiveKit token",
			zap.String("device_id", deviceID),
			zap.String("error", err.Error()),
		)
		response.InternalError(c, "Failed to generate token")
		return
	}

	response.Success(c, gin.H{
		"token":      token,
		"room_id":    deviceID,
		"server_url": h.liveKitConfig.PublicServerURL,
	})
}

// GetDeviceLiveKitToken 生成LiveKit连接Token (Android设备专用)
// GET /api/v1/device/:id/livekit-token (需要设备认证)
func (h *DeviceHandler) GetDeviceLiveKitToken(c *gin.Context) {
	// 从路径参数获取设备ID
	pathDeviceID := c.Param("id")

	// 从上下文获取设备ID（从JWT token）
	deviceID, exists := c.Get("device_id")
	if !exists {
		logger.Warn("LiveKit token request from unauthenticated device",
			zap.String("path_device_id", pathDeviceID),
			zap.String("client_ip", c.ClientIP()),
		)
		response.DeviceError(c, http.StatusUnauthorized, "Device not authenticated")
		return
	}

	deviceIDStr := deviceID.(string)

	// 验证路径参数的设备ID是否匹配认证的设备ID
	if pathDeviceID != deviceIDStr {
		logger.Warn("Device ID mismatch in LiveKit token request",
			zap.String("path_device_id", pathDeviceID),
			zap.String("auth_device_id", deviceIDStr),
		)
		response.DeviceError(c, http.StatusForbidden, "Device ID mismatch")
		return
	}

	logger.Info("Get LiveKit token request for device",
		zap.String("device_id", deviceIDStr),
		zap.String("client_ip", c.ClientIP()),
	)

	// 检查设备是否存在
	_, err := h.deviceService.GetByID(deviceIDStr)
	if err != nil {
		logger.Warn("Device not found",
			zap.String("device_id", deviceIDStr),
		)
		response.DeviceError(c, http.StatusNotFound, "Device not found")
		return
	}

	// 生成LiveKit Token（Android设备用，host角色）
	token, err := h.generateLiveKitToken(deviceIDStr, "host")
	if err != nil {
		logger.Error("Failed to generate LiveKit token",
			zap.String("device_id", deviceIDStr),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	response.DeviceSuccess(c, gin.H{
		"token":      token,
		"room_id":    deviceIDStr,
		"server_url": h.liveKitConfig.PublicServerURL,
	})
}

// GetLiveKitOverview 返回 LiveKit 监控总览（用于首页仪表盘）
// GET /api/v1/web/monitor/livekit-overview
func (h *DeviceHandler) GetLiveKitOverview(c *gin.Context) {
	userID, _ := c.Get("user_id")
	logger.Info("LiveKit overview request received",
		zap.String("path", c.Request.URL.Path),
		zap.String("client_ip", c.ClientIP()),
		zap.Any("user_id", userID),
	)

	degraded := func(reason string) {
		logger.Warn("LiveKit overview degraded", zap.String("reason", reason))
		response.Success(c, gin.H{
			"service_status": "down",
			"error_message":  reason,
			"kpi": gin.H{
				"active_rooms":           0,
				"total_participants":     0,
				"streaming_device_rooms": 0,
				"talkback_sessions":      0,
			},
			"rooms": []gin.H{},
		})
	}

	httpBase, err := liveKitHTTPBaseURL(h.liveKitConfig.ServerURL)
	if err != nil {
		degraded("invalid livekit server url")
		return
	}

	roomService, err := h.newLiveKitRoomService(httpBase, "")
	if err != nil {
		degraded("failed to initialize livekit client")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roomsResp, err := roomService.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		logger.Error("Failed to list livekit rooms", zap.Error(err))
		degraded("failed to query livekit rooms")
		return
	}

	type participantSummary struct {
		Identity string `json:"identity"`
		State    string `json:"state"`
		Tracks   int    `json:"tracks"`
	}

	type roomSummary struct {
		Name         string               `json:"name"`
		CreatedAt    int64                `json:"created_at"`
		DurationSec  int64                `json:"duration_sec"`
		Participants int                  `json:"participants"`
		Publishers   int                  `json:"publishers"`
		HasVideo     bool                 `json:"has_video"`
		HasTalkback  bool                 `json:"has_talkback"`
		Members      []participantSummary `json:"members"`
	}

	overviewRooms := make([]roomSummary, 0, len(roomsResp.GetRooms()))
	var totalParticipants int
	var streamingDeviceRooms int
	var talkbackSessions int

	for _, room := range roomsResp.GetRooms() {
		participantCount := 0
		publishers := 0
		hasVideo := false
		hasTalkback := false
		members := make([]participantSummary, 0)

		roomParticipantService, roomSvcErr := h.newLiveKitRoomService(httpBase, room.GetName())
		if roomSvcErr != nil {
			logger.Warn("Failed to init room participant service, fallback to room-only metrics",
				zap.String("room", room.GetName()),
				zap.Error(roomSvcErr),
			)
			createdAt := room.GetCreationTime()
			duration := int64(0)
			if createdAt > 0 {
				duration = time.Now().Unix() - createdAt
				if duration < 0 {
					duration = 0
				}
			}
			overviewRooms = append(overviewRooms, roomSummary{
				Name:         room.GetName(),
				CreatedAt:    createdAt,
				DurationSec:  duration,
				Participants: participantCount,
				Publishers:   publishers,
				HasVideo:     hasVideo,
				HasTalkback:  hasTalkback,
				Members:      members,
			})
			continue
		}

		participantsResp, pErr := roomParticipantService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
			Room: room.GetName(),
		})
		if pErr != nil {
			logger.Warn("Failed to list participants for room, fallback to room-only metrics",
				zap.String("room", room.GetName()),
				zap.Error(pErr),
			)
		} else {
			participants := participantsResp.GetParticipants()
			participantCount = len(participants)
			totalParticipants += participantCount
			members = make([]participantSummary, 0, len(participants))

			for _, p := range participants {
				if p.GetIsPublisher() {
					publishers++
				}

				for _, t := range p.GetTracks() {
					if t.GetType() == livekit.TrackType_VIDEO && !t.GetMuted() {
						hasVideo = true
					}
					if strings.Contains(strings.ToLower(p.GetIdentity()), "talkback") &&
						t.GetType() == livekit.TrackType_AUDIO &&
						!t.GetMuted() {
						hasTalkback = true
					}
				}

				members = append(members, participantSummary{
					Identity: p.GetIdentity(),
					State:    p.GetState().String(),
					Tracks:   len(p.GetTracks()),
				})
			}
		}

		if hasVideo {
			streamingDeviceRooms++
		}
		if hasTalkback {
			talkbackSessions++
		}

		createdAt := room.GetCreationTime()
		duration := int64(0)
		if createdAt > 0 {
			duration = time.Now().Unix() - createdAt
			if duration < 0 {
				duration = 0
			}
		}

		overviewRooms = append(overviewRooms, roomSummary{
			Name:         room.GetName(),
			CreatedAt:    createdAt,
			DurationSec:  duration,
			Participants: participantCount,
			Publishers:   publishers,
			HasVideo:     hasVideo,
			HasTalkback:  hasTalkback,
			Members:      members,
		})
	}

	response.Success(c, gin.H{
		"service_status": "up",
		"kpi": gin.H{
			"active_rooms":           len(overviewRooms),
			"total_participants":     totalParticipants,
			"streaming_device_rooms": streamingDeviceRooms,
			"talkback_sessions":      talkbackSessions,
		},
		"rooms": overviewRooms,
	})

	logger.Info("LiveKit overview success",
		zap.Int("active_rooms", len(overviewRooms)),
		zap.Int("total_participants", totalParticipants),
		zap.Int("streaming_device_rooms", streamingDeviceRooms),
		zap.Int("talkback_sessions", talkbackSessions),
	)
}

func (h *DeviceHandler) newLiveKitRoomService(baseURL string, roomName string) (livekit.RoomService, error) {
	roomJoin := roomName != ""
	roomList := roomName == ""
	token, err := auth.NewAccessToken(h.liveKitConfig.APIKey, h.liveKitConfig.APISecret).
		SetIdentity("rvcs-monitor").
		SetVideoGrant(&auth.VideoGrant{
			Room:      roomName,
			RoomJoin:  roomJoin,
			RoomList:  roomList,
			RoomAdmin: true,
		}).
		SetValidFor(1 * time.Minute).
		ToJWT()
	if err != nil {
		return nil, err
	}

	insecureSkipVerify := true
	if v := strings.TrimSpace(os.Getenv("LIVEKIT_INSECURE_SKIP_VERIFY")); v != "" {
		insecureSkipVerify = strings.EqualFold(v, "1") ||
			strings.EqualFold(v, "true") ||
			strings.EqualFold(v, "yes") ||
			strings.EqualFold(v, "y")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify}, // self-signed cert support
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &authHeaderTransport{
			base:  transport,
			token: token,
		},
	}
	return livekit.NewRoomServiceProtobufClient(baseURL, client), nil
}

func liveKitHTTPBaseURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid url: %s", raw)
	}
	switch strings.ToLower(u.Scheme) {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	}
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

type authHeaderTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	r.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(r)
}

// generateLiveKitToken 生成LiveKit Token
func (h *DeviceHandler) generateLiveKitToken(deviceID, role string) (string, error) {
	// 创建访问令牌
	at := auth.NewAccessToken(h.liveKitConfig.APIKey, h.liveKitConfig.APISecret)

	// 设置参与者权限和身份
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     deviceID,
	}

	// 根据角色设置不同的权限
	switch role {
	case "host":
		canPublish, canSubscribe := true, true
		grant.CanPublish = &canPublish
		grant.CanSubscribe = &canSubscribe
		at.SetIdentity(deviceID + "-host")
	case "talkback":
		canPublish, canSubscribe := true, true
		grant.CanPublish = &canPublish
		grant.CanSubscribe = &canSubscribe
		at.SetIdentity(deviceID + "-talkback")
	case "viewer":
		canPublish, canSubscribe := false, true
		grant.CanPublish = &canPublish
		grant.CanSubscribe = &canSubscribe
		at.SetIdentity(deviceID + "-viewer")
	default:
		canPublish, canSubscribe := false, true
		grant.CanPublish = &canPublish
		grant.CanSubscribe = &canSubscribe
		at.SetIdentity(deviceID)
	}

	at.AddGrant(grant)

	// 设置过期时间
	at.SetValidFor(h.liveKitConfig.TokenExpiry)

	// 生成JWT token
	token, err := at.ToJWT()
	if err != nil {
		return "", err
	}

	logger.Debug("LiveKit token generated",
		zap.String("device_id", deviceID),
		zap.String("role", role),
		zap.Duration("expiry", h.liveKitConfig.TokenExpiry),
	)

	return token, nil
}

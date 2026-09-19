package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"rvcs/internal/dto"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/repository"
	"rvcs/internal/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DeviceService interface {
	Register(req *dto.DeviceRegisterRequest) (*model.Device, error)
	FindByNameAndModel(name, deviceModel string) (*model.Device, error)             // 根据名称和型号查找设备
	Create(req *dto.DeviceRegisterRequest) (*model.Device, error)                   // 管理员创建设备
	RegisterWithCode(req *RegistrationWithCodeRequest) (*RegistrationResult, error) // 使用注册码注册
	Heartbeat(deviceID string, req *dto.DeviceHeartbeatRequest, clientIP string) error
	GetByID(id string) (*model.Device, error)
	List(page, pageSize int, status string) (*dto.DeviceListResponse, error)
	Update(id string, req *dto.DeviceUpdateRequest) (*model.Device, error)
	Delete(id string) error
	Control(id string, req *dto.DeviceControlRequest) error
	GetPendingCommands(deviceID string, limit int) ([]*model.DeviceCommand, error)
	AckCommand(deviceID string, commandID uint, success bool, result string) error
	GetLogs(id string, page, pageSize int) (*dto.ActivityLogListResponse, error)
	GetAllLogs(page, pageSize int) (*dto.ActivityLogListResponse, error)
}

type deviceService struct {
	deviceRepo           repository.DeviceRepository
	sessionRepo          repository.DeviceSessionRepository
	logRepo              repository.ActivityLogRepository
	commandRepo          repository.DeviceCommandRepository
	registrationCodeRepo repository.RegistrationCodeRepository
	hub                  *websocket.Hub
	jwtManager           any // JWT manager interface
}

func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	sessionRepo repository.DeviceSessionRepository,
	logRepo repository.ActivityLogRepository,
	commandRepo repository.DeviceCommandRepository,
	registrationCodeRepo repository.RegistrationCodeRepository,
	hub *websocket.Hub,
) DeviceService {
	return &deviceService{
		deviceRepo:           deviceRepo,
		sessionRepo:          sessionRepo,
		logRepo:              logRepo,
		commandRepo:          commandRepo,
		registrationCodeRepo: registrationCodeRepo,
		hub:                  hub,
	}
}

func (s *deviceService) Register(req *dto.DeviceRegisterRequest) (*model.Device, error) {
	// 生成设备唯一标识
	deviceID := uuid.New().String()

	// 生成设备令牌
	token := uuid.New().String()

	device := &model.Device{
		ID:             deviceID,
		Name:           req.Name,
		IPAddress:      stringPtr(req.IPAddress),
		Manufacturer:   stringPtr(req.Manufacturer),
		AndroidVersion: stringPtr(req.AndroidVersion),
		AppVersion:     stringPtr(req.AppVersion),
		Token:          &token,
		Status:         model.DeviceStatusOffline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.deviceRepo.Create(device)
	if err != nil {
		return nil, err
	}

	// 记录活动日志
	log := &model.DeviceActivityLog{
		DeviceID:  deviceID,
		Action:    "register",
		Details:   stringToJSON("Device registered"),
		CreatedAt: time.Now(),
	}
	s.logRepo.Create(log)

	return device, nil
}

// FindByNameAndModel 根据设备名称和型号查找设备
func (s *deviceService) FindByNameAndModel(name, deviceModel string) (*model.Device, error) {
	if name == "" || deviceModel == "" {
		return nil, errors.New("name and model are required")
	}
	return s.deviceRepo.FindByNameAndModel(name, deviceModel)
}

func (s *deviceService) Create(req *dto.DeviceRegisterRequest) (*model.Device, error) {
	// 生成设备唯一标识
	deviceID := uuid.New().String()

	// 生成设备令牌
	token := uuid.New().String()

	device := &model.Device{
		ID:             deviceID,
		Name:           req.Name,
		IPAddress:      stringPtr(req.IPAddress),
		Manufacturer:   stringPtr(req.Manufacturer),
		AndroidVersion: stringPtr(req.AndroidVersion),
		AppVersion:     stringPtr(req.AppVersion),
		Token:          &token,
		Status:         model.DeviceStatusOffline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.deviceRepo.Create(device)
	if err != nil {
		return nil, err
	}

	// 记录活动日志
	log := &model.DeviceActivityLog{
		DeviceID:  deviceID,
		Action:    "create",
		Details:   stringToJSON("Device created by admin"),
		CreatedAt: time.Now(),
	}
	s.logRepo.Create(log)

	return device, nil
}

func (s *deviceService) Heartbeat(deviceID string, req *dto.DeviceHeartbeatRequest, clientIP string) error {
	log.Printf("[SERVICE HEARTBEAT] Processing heartbeat for device: %s, status: %s", deviceID, req.Status)

	// 查找设备
	device, err := s.deviceRepo.FindByID(deviceID)
	if err != nil {
		log.Printf("[SERVICE HEARTBEAT] Device not found: %s, error: %v", deviceID, err)
		return errors.New("device not found")
	}

	// 更新设备状态
	oldStatus := device.Status
	log.Printf("[SERVICE HEARTBEAT] Device %s current status: %s", deviceID, oldStatus)

	device.Status = model.DeviceStatus(req.Status)
	hasBattery := req.Health != nil && req.Health.Battery >= 0 && req.Health.Battery <= 100
	if hasBattery {
		battery := req.Health.Battery
		device.BatteryLevel = &battery
	}
	if clientIP != "" {
		device.IPAddress = &clientIP
	}
	now := time.Now()
	device.LastSeen = &now
	device.UpdatedAt = time.Now()

	err = s.deviceRepo.Update(device)
	if err != nil {
		log.Printf("[SERVICE HEARTBEAT] Failed to update device %s: %v", deviceID, err)
		return err
	}

	log.Printf("[SERVICE HEARTBEAT] Device %s updated successfully, new status: %s", deviceID, device.Status)

	// 记录心跳日志（每次都记录）
	heartbeatDetails := fmt.Sprintf("Heartbeat received, status: %s, previous: %s", req.Status, oldStatus)
	if hasBattery {
		heartbeatDetails = fmt.Sprintf("%s, battery: %d%%", heartbeatDetails, req.Health.Battery)
	}
	var remoteIP *string
	if clientIP != "" {
		remoteIP = stringPtr(clientIP)
	} else if device.IPAddress != nil && *device.IPAddress != "" {
		remoteIP = device.IPAddress
	}
	logEntry := &model.DeviceActivityLog{
		DeviceID:  deviceID,
		Action:    "heartbeat",
		Details:   stringToJSON(heartbeatDetails),
		RemoteIP:  remoteIP,
		CreatedAt: time.Now(),
	}
	err = s.logRepo.Create(logEntry)
	if err != nil {
		log.Printf("[SERVICE HEARTBEAT] Failed to create log entry for device %s: %v", deviceID, err)
	} else {
		log.Printf("[SERVICE HEARTBEAT] Log entry created for device %s", deviceID)
	}

	// 如果状态变化，记录额外的状态变更日志
	if oldStatus != model.DeviceStatus(req.Status) {
		log.Printf("[SERVICE HEARTBEAT] Status changed for device %s: %s -> %s", deviceID, oldStatus, req.Status)
		statusLog := &model.DeviceActivityLog{
			DeviceID:  deviceID,
			Action:    "status_change",
			Details:   stringToJSON("Status changed from " + string(oldStatus) + " to " + req.Status),
			CreatedAt: time.Now(),
		}
		s.logRepo.Create(statusLog)

		// 广播状态变化
		s.hub.Broadcast(websocket.Message{
			MsgID:     uuid.New().String(),
			Timestamp: time.Now().Unix(),
			Type:      websocket.MessageType("device_status"),
			Data: map[string]interface{}{
				"device_id": deviceID,
				"status":    req.Status,
			},
		})
		log.Printf("[SERVICE HEARTBEAT] Status change broadcasted for device %s", deviceID)
	}

	log.Printf("[SERVICE HEARTBEAT] Completed processing for device %s", deviceID)
	return nil
}

func (s *deviceService) GetByID(id string) (*model.Device, error) {
	return s.deviceRepo.FindByID(id)
}

func (s *deviceService) List(page, pageSize int, status string) (*dto.DeviceListResponse, error) {
	offset := (page - 1) * pageSize

	devices, total, err := s.deviceRepo.List(offset, pageSize, status)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	deviceList := make([]dto.DeviceInfo, 0, len(devices))
	for _, d := range devices {
		deviceList = append(deviceList, dto.DeviceInfo{
			ID:             d.ID,
			Name:           d.Name,
			Status:         string(d.Status),
			IPAddress:      safeStringPtr(d.IPAddress),
			Manufacturer:   safeStringPtr(d.Manufacturer),
			AndroidVersion: safeStringPtr(d.AndroidVersion),
			AppVersion:     safeStringPtr(d.AppVersion),
			BatteryLevel:   safeIntPtr(d.BatteryLevel),
			LastSeen:       safeTimePtrTime(d.LastSeen),
			CreatedAt:      d.CreatedAt,
		})
	}

	return &dto.DeviceListResponse{
		Total: int(total),
		Items: deviceList,
	}, nil
}

func (s *deviceService) Update(id string, req *dto.DeviceUpdateRequest) (*model.Device, error) {
	device, err := s.deviceRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("device not found")
	}

	// 更新字段
	if req.Name != nil {
		device.Name = *req.Name
	}
	device.UpdatedAt = time.Now()

	err = s.deviceRepo.Update(device)
	if err != nil {
		return nil, err
	}

	// 记录日志
	log := &model.DeviceActivityLog{
		DeviceID:  id,
		Action:    "update",
		Details:   stringToJSON("Device information updated"),
		CreatedAt: time.Now(),
	}
	s.logRepo.Create(log)

	return device, nil
}

// RegistrationWithCodeRequest 使用注册码注册的请求
type RegistrationWithCodeRequest struct {
	RegistrationCode string
	DeviceName       string
	Manufacturer     string
	AndroidVersion   string
	AppVersion       string
	ClientIP         string
	UserAgent        string
}

// RegistrationResult 注册结果
type RegistrationResult struct {
	DeviceID string
}

// RegisterWithCode 使用注册码注册设备
func (s *deviceService) RegisterWithCode(req *RegistrationWithCodeRequest) (*RegistrationResult, error) {
	ctx := context.Background()

	// 验证注册码
	registrationCode, err := s.registrationCodeRepo.GetByCode(ctx, req.RegistrationCode)
	if err != nil {
		return nil, errors.New("invalid registration code")
	}

	// 检查注册码状态
	if registrationCode.Status != "pending" {
		return nil, errors.New("registration code is not available")
	}

	// 检查是否过期
	if time.Now().After(registrationCode.ExpiresAt) {
		return nil, errors.New("registration code has expired")
	}

	// 生成设备ID
	deviceID := fmt.Sprintf("device_%s", uuid.New().String())

	// 创建设备
	device := &model.Device{
		ID:             deviceID,
		Name:           req.DeviceName,
		Manufacturer:   stringPtr(req.Manufacturer),
		AndroidVersion: stringPtr(req.AndroidVersion),
		AppVersion:     stringPtr(req.AppVersion),
		Status:         model.DeviceStatusOffline,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.deviceRepo.Create(device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// 更新注册码状态为已使用
	updates := map[string]interface{}{
		"status":    "used",
		"device_id": deviceID,
		"used_at":   time.Now(),
	}

	if err := s.registrationCodeRepo.Update(ctx, registrationCode.ID, updates); err != nil {
		// 如果更新注册码失败，回滚设备创建
		s.deviceRepo.Delete(deviceID)
		return nil, fmt.Errorf("failed to update registration code: %w", err)
	}

	// 记录活动日志
	log := &model.DeviceActivityLog{
		DeviceID:  deviceID,
		Action:    "register_with_code",
		Details:   stringToJSON(fmt.Sprintf("Device registered with code: %s", req.RegistrationCode)),
		CreatedAt: time.Now(),
	}
	s.logRepo.Create(log)

	return &RegistrationResult{
		DeviceID: deviceID,
	}, nil
}

func (s *deviceService) Delete(id string) error {
	logger.Info("Deleting device", zap.String("device_id", id))

	// 查找设备
	device, err := s.deviceRepo.FindByID(id)
	if err != nil {
		logger.Warn("Device not found", zap.String("device_id", id))
		return errors.New("device not found")
	}
	logger.Info("Device found", zap.String("device_id", id), zap.String("device_name", device.Name))

	// 检查设备是否在线，如果在线则不允许删除
	if device.Status == model.DeviceStatusOnline {
		logger.Warn("Cannot delete online device", zap.String("device_id", id), zap.String("status", string(device.Status)))
		return errors.New("device is online, cannot delete. Please offline the device first")
	}

	// 删除设备
	err = s.deviceRepo.Delete(id)
	if err != nil {
		logger.Error("Failed to delete device from database", zap.String("device_id", id), zap.Error(err))
		return err
	}
	logger.Info("Device deleted from database", zap.String("device_id", id))

	// 删除会话
	err = s.sessionRepo.DeleteByDeviceID(id)
	if err != nil {
		logger.Error("Failed to delete sessions", zap.String("device_id", id), zap.Error(err))
		// 不返回错误，继续执行
	} else {
		logger.Info("Sessions deleted", zap.String("device_id", id))
	}

	// 记录日志
	logEntry := &model.DeviceActivityLog{
		DeviceID:  id,
		Action:    "delete",
		Status:    "success",
		Details:   stringToJSON("Device deleted"),
		CreatedAt: time.Now(),
	}
	err = s.logRepo.Create(logEntry)
	if err != nil {
		logger.Error("Failed to create log entry", zap.String("device_id", id), zap.Error(err))
		// 不返回错误，继续执行
	} else {
		logger.Info("Log entry created", zap.String("device_id", id))
	}

	logger.Info("Device deletion completed", zap.String("device_id", id))

	return nil
}

func (s *deviceService) Control(id string, req *dto.DeviceControlRequest) error {
	// 查找设备
	device, err := s.deviceRepo.FindByID(id)
	if err != nil {
		return errors.New("device not found")
	}

	_ = device

	// 构造控制消息（与 Android 端格式匹配）
	requestID := uuid.New().String()
	data := map[string]interface{}{
		"command":    req.Command,
		"request_id": requestID,
	}

	// 添加额外参数（如分辨率等）
	if len(req.Parameters) > 0 {
		data["data"] = req.Parameters
	}

	message := websocket.Message{
		MsgID:     uuid.New().String(),
		Timestamp: time.Now().Unix(),
		Type:      websocket.MessageType("control"),
		Data:      data,
	}

	payloadBytes, _ := json.Marshal(data)
	cmd := &model.DeviceCommand{
		DeviceID:  id,
		RequestID: requestID,
		Command:   req.Command,
		Payload:   model.JSON(payloadBytes),
		Status:    model.DeviceCommandStatusPending,
		CreatedAt: time.Now(),
	}
	if err := s.commandRepo.Create(cmd); err != nil {
		return err
	}

	// 通过 WebSocket 发送控制命令到指定设备
	sent := s.hub.SendToDevice(id, message)
	if !sent {
		logger.Warn("Device not connected, command queued",
			zap.String("device_id", id),
			zap.String("command", req.Command),
			zap.String("request_id", requestID),
		)
		return nil
	}

	// 记录控制日志
	log := &model.DeviceActivityLog{
		DeviceID:  id,
		Action:    "control",
		Details:   stringToJSON("Command: " + req.Command + " (RequestID: " + requestID + ")"),
		CreatedAt: time.Now(),
	}
	s.logRepo.Create(log)

	return nil
}

func (s *deviceService) GetPendingCommands(deviceID string, limit int) ([]*model.DeviceCommand, error) {
	return s.commandRepo.ListPending(deviceID, limit)
}

func (s *deviceService) AckCommand(deviceID string, commandID uint, success bool, result string) error {
	return s.commandRepo.Ack(deviceID, commandID, success, result)
}

func (s *deviceService) GetLogs(id string, page, pageSize int) (*dto.ActivityLogListResponse, error) {
	offset := (page - 1) * pageSize

	logs, total, err := s.logRepo.FindByDeviceID(id, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 获取设备信息（用于填充设备名称）
	device, err := s.deviceRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device info: %w", err)
	}

	// 转换为响应格式
	logList := make([]dto.ActivityLogInfo, 0, len(logs))
	for _, l := range logs {
		logList = append(logList, dto.ActivityLogInfo{
			ID:         uintToString(l.ID),
			DeviceID:   l.DeviceID,
			DeviceName: device.Name,
			Action:     l.Action,
			Details:    string(l.Details),
			RemoteIP:   safeStringPtr(l.RemoteIP),
			CreatedAt:  l.CreatedAt,
		})
	}

	return &dto.ActivityLogListResponse{
		Total: int(total),
		Items: logList,
	}, nil
}

func (s *deviceService) GetAllLogs(page, pageSize int) (*dto.ActivityLogListResponse, error) {
	offset := (page - 1) * pageSize

	logs, total, err := s.logRepo.FindRecent(offset, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	logList := make([]dto.ActivityLogInfo, 0, len(logs))
	for _, l := range logs {
		// 获取设备信息用于填充设备名称
		var deviceName string
		if device, err := s.deviceRepo.FindByID(l.DeviceID); err == nil {
			deviceName = device.Name
		} else {
			deviceName = "Unknown"
		}

		logList = append(logList, dto.ActivityLogInfo{
			ID:         uintToString(l.ID),
			DeviceID:   l.DeviceID,
			DeviceName: deviceName,
			Action:     l.Action,
			Details:    string(l.Details),
			RemoteIP:   safeStringPtr(l.RemoteIP),
			CreatedAt:  l.CreatedAt,
		})
	}

	return &dto.ActivityLogListResponse{
		Total: int(total),
		Items: logList,
	}, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func safeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeTimePtrTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func safeIntPtr(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func uintToString(u uint) string {
	return fmt.Sprintf("%d", u)
}

func stringToJSON(s string) model.JSON {
	return model.JSON([]byte(s))
}

package scheduler

import (
	"time"

	"rvcs/internal/config"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/repository"
	"rvcs/internal/websocket"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OfflineChecker 设备离线检测器
type OfflineChecker struct {
	deviceRepo repository.DeviceRepository
	hub        *websocket.Hub
	cfg        *config.Config
	ticker     *time.Ticker
	stopChan   chan struct{}
}

// NewOfflineChecker 创建离线检测器
func NewOfflineChecker(deviceRepo repository.DeviceRepository, hub *websocket.Hub, cfg *config.Config) *OfflineChecker {
	return &OfflineChecker{
		deviceRepo: deviceRepo,
		hub:        hub,
		cfg:        cfg,
		stopChan:   make(chan struct{}),
	}
}

// Start 启动离线检测
func (oc *OfflineChecker) Start() {
	interval := oc.cfg.WebSocket.OfflineCheckInterval
	if interval == 0 {
		interval = 60 * time.Second // 默认60秒
	}

	oc.ticker = time.NewTicker(interval)

	go func() {
		logger.Info("Offline checker started",
			zap.Duration("check_interval", interval),
			zap.Duration("offline_timeout", oc.cfg.WebSocket.OfflineTimeout),
		)

		for {
			select {
			case <-oc.ticker.C:
				oc.checkOfflineDevices()
			case <-oc.stopChan:
				logger.Info("Offline checker stopped")
				return
			}
		}
	}()
}

// Stop 停止离线检测
func (oc *OfflineChecker) Stop() {
	if oc.ticker != nil {
		oc.ticker.Stop()
	}
	close(oc.stopChan)
}

// checkOfflineDevices 检测离线设备
func (oc *OfflineChecker) checkOfflineDevices() {
	devices, err := oc.deviceRepo.GetOnlineDevices()
	if err != nil {
		logger.Error("Failed to get online devices for offline check", zap.Error(err))
		return
	}

	timeout := time.Now().Add(-oc.cfg.WebSocket.OfflineTimeout)
	if oc.cfg.WebSocket.OfflineTimeout == 0 {
		timeout = time.Now().Add(-90 * time.Second) // 默认90秒
	}

	offlineCount := 0

	for _, device := range devices {
		if device.LastSeen != nil && device.LastSeen.Before(timeout) {
			// 设备超时，设置为离线
			oldStatus := device.Status
			device.Status = model.DeviceStatusOffline
			if err := oc.deviceRepo.Update(device); err != nil {
				logger.Error("Failed to update device status to offline",
					zap.String("device_id", device.ID),
					zap.String("device_name", device.Name),
					zap.Error(err),
				)
				continue
			}

			logger.Info("Device automatically set to offline",
				zap.String("device_id", device.ID),
				zap.String("device_name", device.Name),
				zap.String("old_status", string(oldStatus)),
				zap.Time("last_seen", *device.LastSeen),
				zap.Duration("timeout_duration", oc.cfg.WebSocket.OfflineTimeout),
			)

			// 通过 WebSocket 广播状态变更
			oc.broadcastDeviceStatus(device)
			offlineCount++
		}
	}

	if offlineCount > 0 {
		logger.Debug("Offline check completed",
			zap.Int("devices_checked", len(devices)),
			zap.Int("offline_devices", offlineCount),
		)
	}
}

// broadcastDeviceStatus 广播设备状态变更
func (oc *OfflineChecker) broadcastDeviceStatus(device *model.Device) {
	message := websocket.Message{
		MsgID:     uuid.New().String(),
		Timestamp: time.Now().Unix(),
		Type:      websocket.MessageType("device_status_changed"),
		Data: map[string]interface{}{
			"device_id": device.ID,
			"status":    device.Status,
		},
	}

	oc.hub.Broadcast(message)
	logger.Debug("Device status change broadcasted",
		zap.String("device_id", device.ID),
		zap.String("status", string(device.Status)),
	)
}

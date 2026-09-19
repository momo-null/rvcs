package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"rvcs/internal/api/middleware"
	"rvcs/internal/api/response"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/service"
)

type AlertHandler struct {
	alertService *service.AlertService
}

func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
	}
}

// GetAlerts 获取告警列表
// @Summary 获取告警列表
// @Tags Alert
// @Accept json
// @Produce json
// @Param device_id query string false "设备ID"
// @Param level query string false "告警级别"
// @Param acknowledged query bool false "是否已确认"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response
// @Router /api/v1/alerts [get]
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	deviceID := c.Query("device_id")
	level := c.Query("level")
	acknowledged := c.Query("acknowledged")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	var ack *bool
	if acknowledged == "true" {
		ack = boolPtr(true)
	} else if acknowledged == "false" {
		ack = boolPtr(false)
	}

	logger.Debug("Get alerts request",
		zap.String("device_id", deviceID),
		zap.String("level", level),
		zap.Any("acknowledged", ack),
		zap.String("page", page),
		zap.String("page_size", pageSize),
	)

	alerts, total, err := h.alertService.GetAlerts(deviceID, level, ack, parsePage(page), parsePageSize(pageSize))
	if err != nil {
		logger.Error("Failed to get alerts", zap.Error(err))
		response.InternalError(c, "Failed to get alerts")
		return
	}

	response.Success(c, gin.H{
		"alerts":    alerts,
		"total":     total,
		"page":      parsePage(page),
		"page_size": parsePageSize(pageSize),
	})
}

// AcknowledgeAlert 确认告警
// @Summary 确认告警
// @Tags Alert
// @Accept json
// @Produce json
// @Param alert_id path string true "告警ID"
// @Success 200 {object} Response
// @Router /api/v1/alerts/{alert_id}/acknowledge [post]
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("alert_id")
	userID := c.GetString("user_id")

	logger.Info("Acknowledge alert request",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
	)

	if err := h.alertService.AcknowledgeAlert(alertID, userID); err != nil {
		logger.Error("Failed to acknowledge alert",
			zap.Error(err),
			zap.String("alert_id", alertID),
			zap.String("user_id", userID),
		)
		response.InternalError(c, "Failed to acknowledge alert")
		return
	}

	logger.Info("Alert acknowledged",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
	)

	response.Success(c, gin.H{
		"alert_id":     alertID,
		"acknowledged": true,
	})
}

// ResolveAlert 解决告警
// @Summary 解决告警
// @Tags Alert
// @Accept json
// @Produce json
// @Param alert_id path string true "告警ID"
// @Success 200 {object} Response
// @Router /api/v1/alerts/{alert_id}/resolve [post]
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("alert_id")
	userID := c.GetString("user_id")

	logger.Info("Resolve alert request",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
	)

	if err := h.alertService.ResolveAlert(alertID, userID); err != nil {
		logger.Error("Failed to resolve alert",
			zap.Error(err),
			zap.String("alert_id", alertID),
			zap.String("user_id", userID),
		)
		response.InternalError(c, "Failed to resolve alert")
		return
	}

	logger.Info("Alert resolved",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID),
	)

	response.Success(c, gin.H{
		"alert_id": alertID,
		"resolved": true,
	})
}

// GetAlertStats 获取告警统计
// @Summary 获取告警统计
// @Tags Alert
// @Accept json
// @Produce json
// @Param period query int false "统计周期（小时）" default(24)
// @Success 200 {object} Response
// @Router /api/v1/alerts/stats [get]
func (h *AlertHandler) GetAlertStats(c *gin.Context) {
	periodHours := c.DefaultQuery("period", "24")
	period := time.Duration(parsePage(periodHours)) * time.Hour

	logger.Debug("Get alert stats request", zap.String("period_hours", periodHours))

	stats, err := h.alertService.GetAlertStats(period)
	if err != nil {
		logger.Error("Failed to get alert stats", zap.Error(err))
		response.InternalError(c, "Failed to get alert stats")
		return
	}

	response.Success(c, stats)
}

// EvaluateRules 触发规则评估（内部使用）
// @Summary 触发规则评估
// @Tags Alert
// @Accept json
// @Produce json
// @Router /api/v1/alerts/evaluate [post]
func (h *AlertHandler) EvaluateRules(c *gin.Context) {
	var event service.AlertEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Debug("Evaluate alert rules request",
		zap.String("event_type", event.EventType),
		zap.String("device_id", event.DeviceID),
	)

	if err := h.alertService.EvaluateRules(&event); err != nil {
		logger.Error("Failed to evaluate alert rules",
			zap.Error(err),
			zap.String("event_type", event.EventType),
			zap.String("device_id", event.DeviceID),
		)
		response.InternalError(c, "Failed to evaluate rules")
		return
	}

	response.Success(c, "Rules evaluated")
}

type deviceAlertRequest struct {
	EventType string                 `json:"event_type" binding:"required"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

// ReportDeviceAlert allows device clients to create direct alerts via device auth.
func (h *AlertHandler) ReportDeviceAlert(c *gin.Context) {
	var req deviceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	deviceID := c.GetString(middleware.DeviceContextKey)
	if deviceID == "" {
		response.Unauthorized(c, "invalid device context")
		return
	}

	level := model.AlertLevelWarning
	switch req.Level {
	case string(model.AlertLevelInfo):
		level = model.AlertLevelInfo
	case string(model.AlertLevelWarning), "":
		level = model.AlertLevelWarning
	case string(model.AlertLevelError):
		level = model.AlertLevelError
	case string(model.AlertLevelCritical):
		level = model.AlertLevelCritical
	default:
		response.BadRequest(c, "Invalid level")
		return
	}

	event := &service.AlertEvent{
		DeviceID:  deviceID,
		EventType: req.EventType,
		Level:     level,
		Message:   req.Message,
		Data:      req.Data,
		Timestamp: time.Now(),
	}

	alert, err := h.alertService.CreateAlert(event)
	if err != nil {
		logger.Error("Failed to create device alert",
			zap.Error(err),
			zap.String("device_id", deviceID),
			zap.String("event_type", req.EventType),
		)
		response.InternalError(c, "Failed to create alert")
		return
	}

	response.Success(c, gin.H{
		"alert_id":   alert.ID,
		"device_id":  alert.DeviceID,
		"event_type": alert.EventType,
		"level":      alert.Level,
	})
}

// Helper functions
func parsePage(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil || i < 1 {
		return 1
	}
	return i
}

func parsePageSize(s string) int {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil || i < 1 || i > 100 {
		return 20
	}
	return i
}

func boolPtr(b bool) *bool {
	return &b
}

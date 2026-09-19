package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/repository"
)

type AlertService struct {
	repo           repository.AlertRepository
	ruleRepo       repository.AlertRuleRepository
	smtpHost       string
	smtpPort       int
	smtpUser       string
	smtpPassword   string
	smtpFrom       string
	webhookEnabled bool
}

// NewAlertService 创建告警服务
func NewAlertService(repo repository.AlertRepository, ruleRepo repository.AlertRuleRepository) *AlertService {
	return &AlertService{
		repo:         repo,
		ruleRepo:     ruleRepo,
		smtpHost:     "smtp.gmail.com",
		smtpPort:     587,
		smtpUser:     "", // 从配置读取
		smtpPassword: "", // 从配置读取
		smtpFrom:     "", // 从配置读取
	}
}

// AlertEvent 告警事件
type AlertEvent struct {
	DeviceID   string
	EventType  string
	Level      model.AlertLevel
	Message    string
	Data       map[string]interface{}
	Timestamp  time.Time
}

// CreateAlert 创建告警
func (s *AlertService) CreateAlert(event *AlertEvent) (*model.Alert, error) {
	alert := &model.Alert{
		ID:          uuid.New().String(),
		DeviceID:    event.DeviceID,
		EventType:   event.EventType,
		Level:       event.Level,
		Message:     &event.Message,
		Acknowledged: false,
		CreatedAt:   time.Now(),
	}

	// 序列化数据
	if event.Data != nil {
		dataBytes, _ := json.Marshal(event.Data)
		alert.Data = model.JSON(dataBytes)
	}

	if err := s.repo.Create(alert); err != nil {
		logger.Error("Failed to create alert", zap.Error(err))
		return nil, err
	}

	logger.Warn("Alert created",
		zap.String("alert_id", alert.ID),
		zap.String("device_id", alert.DeviceID),
		zap.String("event_type", alert.EventType),
		zap.String("level", string(alert.Level)),
	)

	return alert, nil
}

// EvaluateRules 评估告警规则
func (s *AlertService) EvaluateRules(event *AlertEvent) error {
	// 获取该事件类型的所有启用规则
	rules, err := s.ruleRepo.GetEnabledRulesByEventType(event.EventType)
	if err != nil {
		logger.Error("Failed to get alert rules", zap.Error(err))
		return err
	}

	// 评估每条规则
	for _, rule := range rules {
		if s.matchConditions(rule.Conditions, event) {
			// 创建告警
			alert := &model.Alert{
				ID:          uuid.New().String(),
				DeviceID:    event.DeviceID,
				RuleID:      &rule.ID,
				EventType:   event.EventType,
				Level:       rule.Level,
				Message:     &rule.Name,
				Acknowledged: false,
				CreatedAt:   time.Now(),
			}

			// 序列化数据
			if event.Data != nil {
				dataBytes, _ := json.Marshal(event.Data)
				alert.Data = model.JSON(dataBytes)
			}

			if err := s.repo.Create(alert); err != nil {
				logger.Error("Failed to create alert from rule", zap.Error(err))
				continue
			}

			logger.Warn("Alert triggered by rule",
				zap.String("rule_id", rule.ID),
				zap.String("alert_id", alert.ID),
				zap.String("device_id", alert.DeviceID),
			)

			// 发送通知
			go s.sendNotifications(alert, rule.NotificationChannels)
		}
	}

	return nil
}

// matchConditions 匹配告警条件
func (s *AlertService) matchConditions(conditions model.JSON, event *AlertEvent) bool {
	// 解析条件
	var condMap map[string]interface{}
	if err := json.Unmarshal(conditions, &condMap); err != nil {
		logger.Error("Failed to parse alert conditions", zap.Error(err))
		return false
	}

	// 简化处理：检查是否匹配
	// 实际应该实现更复杂的条件匹配逻辑
	_, ok := condMap["enabled"]
	if !ok {
		return true // 默认匹配
	}

	enabled, ok := condMap["enabled"].(bool)
	return ok && enabled
}

// sendNotifications 发送告警通知
func (s *AlertService) sendNotifications(alert *model.Alert, channels model.JSON) {
	var channelList []map[string]interface{}
	if err := json.Unmarshal(channels, &channelList); err != nil {
		logger.Error("Failed to parse notification channels", zap.Error(err))
		return
	}

	for _, channel := range channelList {
		channelType, ok := channel["type"].(string)
		if !ok {
			continue
		}

		switch channelType {
		case "email":
			s.sendEmailNotification(alert, channel)
		case "webhook":
			s.sendWebhookNotification(alert, channel)
		case "sms":
			s.sendSMSNotification(alert, channel)
		}
	}
}

// sendEmailNotification 发送邮件通知
func (s *AlertService) sendEmailNotification(alert *model.Alert, channel map[string]interface{}) {
	recipients, ok := channel["recipients"].([]interface{})
	if !ok || len(recipients) == 0 {
		logger.Warn("No email recipients configured")
		return
	}

	// 构造邮件内容
	_ = fmt.Sprintf("[%s] RVCS Alert: %s", alert.Level, alert.EventType)
	_ = fmt.Sprintf(`
		<h2>RVCS Alert Notification</h2>
		<p><strong>Device ID:</strong> %s</p>
		<p><strong>Event Type:</strong> %s</p>
		<p><strong>Level:</strong> %s</p>
		<p><strong>Message:</strong> %s</p>
		<p><strong>Time:</strong> %s</p>
	`,
		alert.DeviceID,
		alert.EventType,
		alert.Level,
		safeString(alert.Message),
		alert.CreatedAt.Format(time.RFC3339),
	)

	// 简化处理，实际需要配置 SMTP
	logger.Info("Email notification would be sent",
		zap.String("alert_id", alert.ID),
		zap.Int("recipients", len(recipients)),
	)

	// 实际发送邮件（需要配置 SMTP）
	// if err := s.sendEmail(recipients, subject, body); err != nil {
	//     logger.Error("Failed to send email", zap.Error(err))
	// }
}

// sendWebhookNotification 发送 Webhook 通知
func (s *AlertService) sendWebhookNotification(alert *model.Alert, channel map[string]interface{}) {
	webhookURL, ok := channel["url"].(string)
	if !ok || webhookURL == "" {
		logger.Warn("No webhook URL configured")
		return
	}

	// 构造 payload
	_ = map[string]interface{}{
		"alert_id":    alert.ID,
		"device_id":   alert.DeviceID,
		"event_type":  alert.EventType,
		"level":       alert.Level,
		"message":     alert.Message,
		"data":        alert.Data,
		"created_at":  alert.CreatedAt.Format(time.RFC3339),
	}

	// 发送 HTTP 请求（简化处理）
	logger.Info("Webhook notification would be sent",
		zap.String("alert_id", alert.ID),
		zap.String("url", webhookURL),
	)

	// 实际发送 HTTP POST 请求
	// if err := s.sendHTTPPost(webhookURL, payload); err != nil {
	//     logger.Error("Failed to send webhook", zap.Error(err))
	// }
}

// sendSMSNotification 发送短信通知
func (s *AlertService) sendSMSNotification(alert *model.Alert, channel map[string]interface{}) {
	phoneNumbers, ok := channel["phone_numbers"].([]interface{})
	if !ok || len(phoneNumbers) == 0 {
		logger.Warn("No phone numbers configured")
		return
	}

	logger.Info("SMS notification would be sent",
		zap.String("alert_id", alert.ID),
		zap.Int("recipients", len(phoneNumbers)),
	)

	// 实际需要接入短信服务（如阿里云短信、腾讯云短信等）
}

// AcknowledgeAlert 确认告警
func (s *AlertService) AcknowledgeAlert(alertID, userID string) error {
	alert, err := s.repo.GetByID(alertID)
	if err != nil {
		return err
	}

	alert.Acknowledged = true
	alert.AcknowledgedBy = &userID
	now := time.Now()
	alert.AcknowledgedAt = &now

	return s.repo.Update(alert)
}

// ResolveAlert 解决告警
func (s *AlertService) ResolveAlert(alertID, userID string) error {
	// 先确认告警
	if err := s.AcknowledgeAlert(alertID, userID); err != nil {
		return err
	}

	// 可以添加解决时间字段
	return nil
}

// GetAlerts 获取告警列表
func (s *AlertService) GetAlerts(deviceID, level string, acknowledged *bool, page, pageSize int) ([]*model.Alert, int64, error) {
	return s.repo.GetAlerts(deviceID, level, acknowledged, page, pageSize)
}

// GetAlertStats 获取告警统计
func (s *AlertService) GetAlertStats(period time.Duration) (*AlertStats, error) {
	since := time.Now().Add(-period)

	total, err := s.repo.CountAlerts("", "", nil, since)
	if err != nil {
		return nil, err
	}

	unacknowledged, err := s.repo.CountAlerts("", "", boolPtr(false), since)
	if err != nil {
		return nil, err
	}

	acknowledged, err := s.repo.CountAlerts("", "", boolPtr(true), since)
	if err != nil {
		return nil, err
	}

	// 按级别统计
	criticalCount, _ := s.repo.CountAlerts("", string(model.AlertLevelCritical), nil, since)
	errorCount, _ := s.repo.CountAlerts("", string(model.AlertLevelError), nil, since)
	warningCount, _ := s.repo.CountAlerts("", string(model.AlertLevelWarning), nil, since)

	return &AlertStats{
		Total:         total,
		Unacknowledged: unacknowledged,
		Acknowledged:   acknowledged,
		ByLevel: map[string]int64{
			"critical": criticalCount,
			"error":    errorCount,
			"warning":  warningCount,
			"info":     0, // 计算得到
		},
	}, nil
}

// AlertStats 告警统计
type AlertStats struct {
	Total         int64           `json:"total"`
	Unacknowledged int64          `json:"unacknowledged"`
	Acknowledged   int64          `json:"acknowledged"`
	ByLevel       map[string]int64 `json:"by_level"`
	ByDevice      []DeviceAlertStats `json:"by_device"`
}

// DeviceAlertStats 设备告警统计
type DeviceAlertStats struct {
	DeviceID   string `json:"device_id"`
	AlertCount int64  `json:"alert_count"`
}

// Helper functions
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolPtr(b bool) *bool {
	return &b
}

package repository

import (
	"rvcs/internal/model"
	"time"

	"gorm.io/gorm"
)

type AlertRepository interface {
	Create(alert *model.Alert) error
	Update(alert *model.Alert) error
	GetByID(id string) (*model.Alert, error)
	GetAlerts(deviceID, level string, acknowledged *bool, page, pageSize int) ([]*model.Alert, int64, error)
	CountAlerts(deviceID, level string, acknowledged *bool, since time.Time) (int64, error)
	GetUnacknowledgedAlerts() ([]*model.Alert, error)
	DeleteOldAlerts(before time.Time) error
}

type AlertRuleRepository interface {
	Create(rule *model.AlertRule) error
	Update(rule *model.AlertRule) error
	GetByID(id string) (*model.AlertRule, error)
	GetEnabledRulesByEventType(eventType string) ([]*model.AlertRule, error)
	GetRulesByUser(userID string) ([]*model.AlertRule, error)
	Delete(id string) error
}

type alertRepository struct {
	db *gorm.DB
}

type alertRuleRepository struct {
	db *gorm.DB
}

func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{db: db}
}

func NewAlertRuleRepository(db *gorm.DB) AlertRuleRepository {
	return &alertRuleRepository{db: db}
}

// Alert Repository Methods
func (r *alertRepository) Create(alert *model.Alert) error {
	return r.db.Create(alert).Error
}

func (r *alertRepository) Update(alert *model.Alert) error {
	return r.db.Save(alert).Error
}

func (r *alertRepository) GetByID(id string) (*model.Alert, error) {
	var alert model.Alert
	err := r.db.Where("id = ?", id).First(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

func (r *alertRepository) GetAlerts(deviceID, level string, acknowledged *bool, page, pageSize int) ([]*model.Alert, int64, error) {
	var alerts []*model.Alert
	var total int64

	query := r.db.Model(&model.Alert{})

	// 筛选条件
	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if acknowledged != nil {
		query = query.Where("acknowledged = ?", *acknowledged)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&alerts).Error

	return alerts, total, err
}

func (r *alertRepository) CountAlerts(deviceID, level string, acknowledged *bool, since time.Time) (int64, error) {
	var count int64
	query := r.db.Model(&model.Alert{})

	if deviceID != "" {
		query = query.Where("device_id = ?", deviceID)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if acknowledged != nil {
		query = query.Where("acknowledged = ?", *acknowledged)
	}
	if !since.IsZero() {
		query = query.Where("created_at >= ?", since)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *alertRepository) GetUnacknowledgedAlerts() ([]*model.Alert, error) {
	var alerts []*model.Alert
	err := r.db.Where("acknowledged = ?", false).
		Order("created_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (r *alertRepository) DeleteOldAlerts(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&model.Alert{}).Error
}

// AlertRule Repository Methods
func (r *alertRuleRepository) Create(rule *model.AlertRule) error {
	return r.db.Create(rule).Error
}

func (r *alertRuleRepository) Update(rule *model.AlertRule) error {
	return r.db.Save(rule).Error
}

func (r *alertRuleRepository) GetByID(id string) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alertRuleRepository) GetEnabledRulesByEventType(eventType string) ([]*model.AlertRule, error) {
	var rules []*model.AlertRule
	err := r.db.Where("event_type = ? AND enabled = ?", eventType, true).
		Find(&rules).Error
	return rules, err
}

func (r *alertRuleRepository) GetRulesByUser(userID string) ([]*model.AlertRule, error) {
	var rules []*model.AlertRule
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&rules).Error
	return rules, err
}

func (r *alertRuleRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.AlertRule{}).Error
}

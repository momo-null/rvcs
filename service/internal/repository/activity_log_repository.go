package repository

import (
	"rvcs/internal/model"

	"gorm.io/gorm"
)

type ActivityLogRepository interface {
	Create(log *model.DeviceActivityLog) error
	FindByDeviceID(deviceID string, offset, limit int) ([]*model.DeviceActivityLog, int64, error)
	FindByUserID(userID string, offset, limit int) ([]*model.DeviceActivityLog, int64, error)
	FindRecent(offset, limit int) ([]*model.DeviceActivityLog, int64, error)
}

type activityLogRepository struct {
	db *gorm.DB
}

func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

func (r *activityLogRepository) Create(log *model.DeviceActivityLog) error {
	return r.db.Create(log).Error
}

func (r *activityLogRepository) FindByDeviceID(deviceID string, offset, limit int) ([]*model.DeviceActivityLog, int64, error) {
	var logs []*model.DeviceActivityLog
	var total int64

	err := r.db.Model(&model.DeviceActivityLog{}).
		Where("device_id = ?", deviceID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("device_id = ?", deviceID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	return logs, total, err
}

func (r *activityLogRepository) FindByUserID(userID string, offset, limit int) ([]*model.DeviceActivityLog, int64, error) {
	var logs []*model.DeviceActivityLog
	var total int64

	err := r.db.Model(&model.DeviceActivityLog{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	return logs, total, err
}

func (r *activityLogRepository) FindRecent(offset, limit int) ([]*model.DeviceActivityLog, int64, error) {
	var logs []*model.DeviceActivityLog
	var total int64

	err := r.db.Model(&model.DeviceActivityLog{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error
	return logs, total, err
}

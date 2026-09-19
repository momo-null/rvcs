package repository

import (
	"time"

	"rvcs/internal/model"

	"gorm.io/gorm"
)

type DeviceCommandRepository interface {
	Create(cmd *model.DeviceCommand) error
	ListPending(deviceID string, limit int) ([]*model.DeviceCommand, error)
	Ack(deviceID string, id uint, success bool, result string) error
}

type deviceCommandRepository struct {
	db *gorm.DB
}

func NewDeviceCommandRepository(db *gorm.DB) DeviceCommandRepository {
	return &deviceCommandRepository{db: db}
}

func (r *deviceCommandRepository) Create(cmd *model.DeviceCommand) error {
	return r.db.Create(cmd).Error
}

func (r *deviceCommandRepository) ListPending(deviceID string, limit int) ([]*model.DeviceCommand, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var items []*model.DeviceCommand
	err := r.db.Where("device_id = ? AND status = ?", deviceID, model.DeviceCommandStatusPending).
		Order("id asc").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *deviceCommandRepository) Ack(deviceID string, id uint, success bool, result string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":   model.DeviceCommandStatusAcked,
		"acked_at": &now,
	}
	if result != "" {
		updates["result"] = result
	}
	// success currently only recorded in result text by caller.
	return r.db.Model(&model.DeviceCommand{}).
		Where("id = ? AND device_id = ?", id, deviceID).
		Updates(updates).Error
}

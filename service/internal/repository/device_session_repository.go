package repository

import (
	"rvcs/internal/model"

	"gorm.io/gorm"
)

type DeviceSessionRepository interface {
	Create(session *model.DeviceSession) error
	FindByDeviceID(deviceID string) (*model.DeviceSession, error)
	Update(session *model.DeviceSession) error
	DeleteByDeviceID(deviceID string) error
	DeleteExpired() error
}

type deviceSessionRepository struct {
	db *gorm.DB
}

func NewDeviceSessionRepository(db *gorm.DB) DeviceSessionRepository {
	return &deviceSessionRepository{db: db}
}

func (r *deviceSessionRepository) Create(session *model.DeviceSession) error {
	return r.db.Create(session).Error
}

func (r *deviceSessionRepository) FindByDeviceID(deviceID string) (*model.DeviceSession, error) {
	var session model.DeviceSession
	err := r.db.Where("device_id = ?", deviceID).Order("created_at DESC").First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *deviceSessionRepository) Update(session *model.DeviceSession) error {
	return r.db.Save(session).Error
}

func (r *deviceSessionRepository) DeleteByDeviceID(deviceID string) error {
	return r.db.Where("device_id = ?", deviceID).Delete(&model.DeviceSession{}).Error
}

func (r *deviceSessionRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", "NOW()").Delete(&model.DeviceSession{}).Error
}

package repository

import (
	"rvcs/internal/model"

	"gorm.io/gorm"
)

type DeviceRepository interface {
	Create(device *model.Device) error
	FindByID(id string) (*model.Device, error)
	FindByToken(token string) (*model.Device, error)
	FindByNameAndModel(name, deviceModel string) (*model.Device, error) // 根据名称和型号查找
	Update(device *model.Device) error
	Delete(id string) error
	List(offset, limit int, status string) ([]*model.Device, int64, error)
	FindByGroupID(groupID string, offset, limit int) ([]*model.Device, int64, error)
	UpdateStatus(id, status string) error
	GetOnlineDevices() ([]*model.Device, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(device *model.Device) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) FindByID(id string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) FindByToken(token string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("token = ?", token).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) FindByUUID(uuid string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("device_uuid = ?", uuid).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) FindByNameAndModel(name, deviceModel string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("name = ? AND model = ?", name, deviceModel).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) Update(device *model.Device) error {
	return r.db.Save(device).Error
}

func (r *deviceRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Device{}).Error
}

func (r *deviceRepository) List(offset, limit int, status string) ([]*model.Device, int64, error) {
	var devices []*model.Device
	var total int64

	query := r.db.Model(&model.Device{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&devices).Error
	return devices, total, err
}

func (r *deviceRepository) FindByGroupID(groupID string, offset, limit int) ([]*model.Device, int64, error) {
	var devices []*model.Device
	var total int64

	err := r.db.Model(&model.Device{}).
		Joins("JOIN device_group_members ON device_group_members.device_id = devices.id").
		Where("device_group_members.group_id = ?", groupID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&devices).Error
	return devices, total, err
}

func (r *deviceRepository) UpdateStatus(id, status string) error {
	return r.db.Model(&model.Device{}).Where("id = ?", id).Update("status", status).Error
}

func (r *deviceRepository) GetOnlineDevices() ([]*model.Device, error) {
	var devices []*model.Device
	err := r.db.Where("status = ?", "online").Find(&devices).Error
	return devices, err
}

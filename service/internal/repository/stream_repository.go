package repository

import (
	"rvcs/internal/model"

	"gorm.io/gorm"
)

type StreamRepository interface {
	Create(stream *model.Stream) error
	Update(stream *model.Stream) error
	GetByID(id string) (*model.Stream, error)
	GetByDevice(deviceID string) ([]*model.Stream, error)
	GetActiveStreams() ([]*model.Stream, error)
	GetByUser(userID string) ([]*model.Stream, error)
	Delete(id string) error
}

type streamRepository struct {
	db *gorm.DB
}

func NewStreamRepository(db *gorm.DB) StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) Create(stream *model.Stream) error {
	return r.db.Create(stream).Error
}

func (r *streamRepository) Update(stream *model.Stream) error {
	return r.db.Save(stream).Error
}

func (r *streamRepository) GetByID(id string) (*model.Stream, error) {
	var stream model.Stream
	err := r.db.Where("id = ?", id).First(&stream).Error
	if err != nil {
		return nil, err
	}
	return &stream, nil
}

func (r *streamRepository) GetByDevice(deviceID string) ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("device_id = ?", deviceID).Order("created_at DESC").Find(&streams).Error
	return streams, err
}

func (r *streamRepository) GetActiveStreams() ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("status = ?", model.StreamStatusActive).
		Order("created_at DESC").
		Find(&streams).Error
	return streams, err
}

func (r *streamRepository) GetByUser(userID string) ([]*model.Stream, error) {
	var streams []*model.Stream
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&streams).Error
	return streams, err
}

func (r *streamRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Stream{}).Error
}

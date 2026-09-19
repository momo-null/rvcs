package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/repository"
)

type StreamService struct {
	repo   repository.StreamRepository
	srsAPI string // SRS HTTP API 地址
}

type CreateStreamRequest struct {
	DeviceID   string `json:"device_id"`
	UserID     string `json:"user_id"`
	StreamType string `json:"stream_type"` // webrtc
	Quality    string `json:"quality"`    // low, medium, high
	Duration   int    `json:"duration"`   // 秒
}

// NewStreamService 创建流服务
func NewStreamService(repo repository.StreamRepository) *StreamService {
	return &StreamService{
		repo:   repo,
		srsAPI: "http://localhost:1985", // SRS 默认 HTTP API 端口
	}
}

// CreateStream 创建新的视频流
func (s *StreamService) CreateStream(req *CreateStreamRequest) (*model.Stream, error) {
	streamID := uuid.New().String()
	streamURL := s.generateStreamURL(streamID, model.StreamType(req.StreamType))
	
	stream := &model.Stream{
		ID:         streamID,
		DeviceID:   req.DeviceID,
		UserID:     req.UserID,
		StreamType: model.StreamType(req.StreamType),
		Quality:    model.StreamQuality(req.Quality),
		Status:     model.StreamStatusActive,
		StartedAt:  timePtr(time.Now()),
		ExpiresAt:  timePtr(time.Now().Add(time.Duration(req.Duration) * time.Second)),
	}

	// 生成流 URL
	stream.StreamURL = &streamURL

	// 生成访问令牌
	stream.AuthToken = stringPtr(generateStreamToken(stream.ID))

	// 保存到数据库
	if err := s.repo.Create(stream); err != nil {
		logger.Error("Failed to save stream", zap.Error(err), zap.String("stream_id", stream.ID))
		return nil, err
	}

	// 通知 SRS 创建流（这里简化处理）
	// 实际需要调用 SRS API 创建推流端点
	logger.Info("Stream created successfully",
		zap.String("stream_id", stream.ID),
		zap.String("device_id", req.DeviceID),
		zap.String("stream_type", req.StreamType),
	)

	return stream, nil
}

// StopStream 停止视频流
func (s *StreamService) StopStream(streamID string) error {
	// 获取流信息
	stream, err := s.repo.GetByID(streamID)
	if err != nil {
		return fmt.Errorf("stream not found: %w", err)
	}

	// 更新状态
	stream.Status = model.StreamStatusInactive

	// 通知 SRS 停止流（这里简化处理）
	logger.Info("Stopping stream",
		zap.String("stream_id", streamID),
		zap.String("device_id", stream.DeviceID),
	)

	// 更新数据库
	if err := s.repo.Update(stream); err != nil {
		return fmt.Errorf("failed to update stream: %w", err)
	}

	return nil
}

// GetStreamByID 根据ID获取流
func (s *StreamService) GetStreamByID(streamID string) (*model.Stream, error) {
	return s.repo.GetByID(streamID)
}

// GetStreamsByDevice 获取设备的所有流
func (s *StreamService) GetStreamsByDevice(deviceID, userID string) ([]*model.Stream, error) {
	// 验证用户是否有权访问此设备的流
	return s.repo.GetByDevice(deviceID)
}

// GetActiveStreams 获取活跃的流
func (s *StreamService) GetActiveStreams(userID string) ([]*model.Stream, error) {
	return s.repo.GetActiveStreams()
}

// generateStreamURL 生成流URL
func (s *StreamService) generateStreamURL(streamID string, streamType model.StreamType) string {
	switch streamType {
	case model.StreamTypeWebRTC:
		return fmt.Sprintf("webrtc://localhost:28443/stream/%s", streamID)
	default:
		return ""
	}
}

// generateStreamToken 生成流访问令牌
func generateStreamToken(streamID string) string {
	// 简化处理，实际应该使用加密签名
	return fmt.Sprintf("stream_token_%s_%d", streamID, time.Now().Unix())
}

// helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

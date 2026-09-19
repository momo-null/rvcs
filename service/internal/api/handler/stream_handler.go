package handler

import (
	"net/http"

	"rvcs/internal/api/response"
	"rvcs/internal/logger"
	"rvcs/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type StreamHandler struct {
	streamService *service.StreamService
}

func NewStreamHandler(streamService *service.StreamService) *StreamHandler {
	return &StreamHandler{
		streamService: streamService,
	}
}

// CreateStream 创建新的视频流
func (h *StreamHandler) CreateStream(c *gin.Context) {
	var req service.CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.DeviceError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	logger.Info("Create stream request",
		zap.String("device_id", req.DeviceID),
		zap.String("stream_type", req.StreamType),
	)

	stream, err := h.streamService.CreateStream(&req)
	if err != nil {
		logger.Warn("Failed to create stream",
			zap.String("device_id", req.DeviceID),
			zap.String("error", err.Error()),
		)
		response.DeviceError(c, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Info("Stream created successfully", zap.String("stream_id", stream.ID))
	response.DeviceCreated(c, stream)
}

// GetStream 获取单个流信息
func (h *StreamHandler) GetStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	logger.Debug("Get stream request", zap.String("stream_id", streamID))

	stream, err := h.streamService.GetStreamByID(streamID)
	if err != nil {
		logger.Warn("Stream not found", zap.String("stream_id", streamID))
		response.NotFound(c, "Stream not found")
		return
	}

	response.Success(c, stream)
}

// GetStreams 获取设备所有流
func (h *StreamHandler) GetStreams(c *gin.Context) {
	deviceID := c.Param("id")

	logger.Debug("Get device streams request", zap.String("device_id", deviceID))

	streams, err := h.streamService.GetStreamsByDevice(deviceID, "")
	if err != nil {
		logger.Error("Failed to get device streams",
			zap.String("device_id", deviceID),
			zap.String("error", err.Error()),
		)
		response.InternalError(c, err.Error())
		return
	}

	response.DeviceSuccess(c, streams)
}

// GetAllStreams 获取所有活跃流
func (h *StreamHandler) GetAllStreams(c *gin.Context) {
	logger.Debug("Get all active streams request")

	streams, err := h.streamService.GetActiveStreams("")
	if err != nil {
		logger.Error("Failed to get all streams", zap.Error(err))
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, streams)
}

// DeleteStream 删除流
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	logger.Info("Delete stream request", zap.String("stream_id", streamID))

	if err := h.streamService.StopStream(streamID); err != nil {
		logger.Warn("Failed to delete stream",
			zap.String("stream_id", streamID),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("Stream deleted successfully", zap.String("stream_id", streamID))
	response.Success(c, gin.H{"message": "Stream deleted successfully"})
}

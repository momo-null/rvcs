package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"rvcs/internal/logger"
	"rvcs/internal/service"
)

type StreamHandler struct {
	streamService *service.StreamService
}

func NewStreamHandler(streamService *service.StreamService) *StreamHandler {
	return &StreamHandler{
		streamService: streamService,
	}
}

// CreateStreamRequest 创建流请求
type CreateStreamRequest struct {
	Action     string `json:"action" binding:"required"`     // start, stop
	StreamType string `json:"stream_type" binding:"required"` // rtsp, rtmp, hls
	Quality    string `json:"quality"`                   // low, medium, high
	Duration   int    `json:"duration"`                  // 流持续时间（秒）
}

// CreateStream 创建视频流
// @Summary 创建视频流
// @Tags Stream
// @Accept json
// @Produce json
// @Param device_id path string true "设备ID"
// @Param request body CreateStreamRequest true "创建流请求"
// @Success 200 {object} Response
// @Router /api/v1/devices/{device_id}/stream [post]
func (h *StreamHandler) CreateStream(c *gin.Context) {
	deviceID := c.Param("device_id")

	var req CreateStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid stream request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request parameters",
		})
		return
	}

	// 验证流类型
	if req.StreamType != "rtsp" && req.StreamType != "rtmp" && req.StreamType != "hls" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid stream type",
		})
		return
	}

	// 验证质量参数
	if req.Quality == "" {
		req.Quality = "medium"
	}
	if req.Quality != "low" && req.Quality != "medium" && req.Quality != "high" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid quality parameter",
		})
		return
	}

	// 设置默认持续时间
	if req.Duration <= 0 {
		req.Duration = 3600 // 1小时
	}

	switch req.Action {
	case "start":
		stream, err := h.streamService.CreateStream(&service.CreateStreamRequest{
			DeviceID:   deviceID,
			UserID:     c.GetString("user_id"),
			StreamType: req.StreamType,
			Quality:    req.Quality,
			Duration:   req.Duration,
		})
		if err != nil {
			logger.Error("Failed to create stream", zap.Error(err), zap.String("device_id", deviceID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to create stream",
			})
			return
		}

		logger.Info("Stream created",
			zap.String("stream_id", stream.ID),
			zap.String("device_id", deviceID),
			zap.String("stream_type", req.StreamType),
		)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"stream_id":  stream.ID,
				"stream_url":  stream.StreamURL,
				"auth_token":  stream.AuthToken,
				"expires_at":  stream.ExpiresAt.Format(time.RFC3339),
				"created_at":  stream.CreatedAt.Format(time.RFC3339),
			},
		})

	case "stop":
		streamID := c.Query("stream_id")
		if streamID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Stream ID is required",
			})
			return
		}

		if err := h.streamService.StopStream(streamID); err != nil {
			logger.Error("Failed to stop stream", zap.Error(err), zap.String("stream_id", streamID))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to stop stream",
			})
			return
		}

		logger.Info("Stream stopped", zap.String("stream_id", streamID))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"stream_id": streamID,
				"status":    "stopped",
			},
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid action. Use 'start' or 'stop'",
		})
	}
}

// GetStreams 获取设备的视频流列表
// @Summary 获取视频流列表
// @Tags Stream
// @Accept json
// @Produce json
// @Param device_id path string true "设备ID"
// @Success 200 {object} Response
// @Router /api/v1/devices/{device_id}/streams [get]
func (h *StreamHandler) GetStreams(c *gin.Context) {
	deviceID := c.Param("device_id")
	userID := c.GetString("user_id")

	streams, err := h.streamService.GetStreamsByDevice(deviceID, userID)
	if err != nil {
		logger.Error("Failed to get streams", zap.Error(err), zap.String("device_id", deviceID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get streams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"device_id": deviceID,
			"streams":   streams,
			"count":     len(streams),
		},
	})
}

// GetStream 获取单个视频流信息
// @Summary 获取视频流信息
// @Tags Stream
// @Accept json
// @Produce json
// @Param stream_id path string true "流ID"
// @Success 200 {object} Response
// @Router /api/v1/streams/{stream_id} [get]
func (h *StreamHandler) GetStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	stream, err := h.streamService.GetStreamByID(streamID)
	if err != nil {
		logger.Error("Stream not found", zap.Error(err), zap.String("stream_id", streamID))
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Stream not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":     stream,
	})
}

// DeleteStream 删除视频流
// @Summary 删除视频流
// @Tags Stream
// @Accept json
// @Produce json
// @Param stream_id path string true "流ID"
// @Success 200 {object} Response
// @Router /api/v1/streams/{stream_id} [delete]
func (h *StreamHandler) DeleteStream(c *gin.Context) {
	streamID := c.Param("stream_id")

	if err := h.streamService.StopStream(streamID); err != nil {
		logger.Error("Failed to delete stream", zap.Error(err), zap.String("stream_id", streamID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete stream",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stream_id": streamID,
		},
	})
}

// GetAllStreams 获取所有活跃的流
// @Summary 获取所有活跃流
// @Tags Stream
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/streams [get]
func (h *StreamHandler) GetAllStreams(c *gin.Context) {
	userID := c.GetString("user_id")

	streams, err := h.streamService.GetActiveStreams(userID)
	if err != nil {
		logger.Error("Failed to get active streams", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get active streams",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"streams": streams,
			"count":   len(streams),
		},
	})
}

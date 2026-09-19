package handler

import (
	"net/http"

	"rvcs/internal/api/response"
	"rvcs/internal/dto"
	"rvcs/internal/logger"
	"rvcs/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	logger.Info("User login attempt",
		zap.String("username", req.Username),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用服务层登录
	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		logger.Warn("User login failed",
			zap.String("username", req.Username),
			zap.String("error", err.Error()),
		)
		response.Unauthorized(c, err.Error())
		return
	}

	logger.Info("User login success",
		zap.String("username", req.Username),
		zap.String("user_id", resp.User.ID),
	)

	response.Success(c, resp)
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("User registration attempt",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("client_ip", c.ClientIP()),
	)

	// 调用服务层注册
	user, err := h.authService.Register(&req)
	if err != nil {
		logger.Warn("User registration failed",
			zap.String("username", req.Username),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User registration success",
		zap.String("username", req.Username),
		zap.String("user_id", user.ID),
	)

	response.Created(c, user)
}

// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	logger.Debug("Token refresh request", zap.String("client_ip", c.ClientIP()))

	// 调用服务层刷新令牌
	resp, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		logger.Warn("Token refresh failed", zap.String("error", err.Error()))
		response.Unauthorized(c, err.Error())
		return
	}

	logger.Info("Token refresh success", zap.String("access_token", resp.AccessToken[:min(len(resp.AccessToken), 20)]+"..."))
	response.Success(c, resp)
}

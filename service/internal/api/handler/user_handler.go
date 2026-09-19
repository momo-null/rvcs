package handler

import (
	"rvcs/internal/api/response"
	"rvcs/internal/dto"
	"rvcs/internal/logger"
	"rvcs/internal/model"
	"rvcs/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUser 获取用户详情
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	logger.Debug("Get user request", zap.String("user_id", id))

	user, err := h.userService.GetUser(id)
	if err != nil {
		logger.Warn("User not found", zap.String("user_id", id))
		response.NotFound(c, "User not found")
		return
	}

	response.Success(c, user)
}

// GetUserList 获取用户列表
func (h *UserHandler) GetUserList(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	logger.Debug("Get user list request",
		zap.Int("page", req.Page),
		zap.Int("page_size", req.PageSize),
		zap.String("role", req.Role),
		zap.String("status", req.Status),
	)

	users, total, err := h.userService.GetUserList(req.Page, req.PageSize, req.Role, req.Status)
	if err != nil {
		logger.Error("Failed to get user list", zap.Error(err))
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"users":     users,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
	})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("Create user request",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("role", req.Role),
		zap.String("status", req.Status),
	)

	// 创建用户对象
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     model.UserRole(req.Role),
		Status:   model.UserStatus(req.Status),
	}

	// 创建用户
	if err := h.userService.CreateUser(user, req.Password); err != nil {
		logger.Warn("Failed to create user",
			zap.String("username", req.Username),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User created successfully", zap.String("user_id", user.ID))
	response.Created(c, user)
}

// UpdateUser 更新用户
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("Update user request",
		zap.String("user_id", id),
		zap.Any("username", req.Username),
		zap.Any("email", req.Email),
		zap.Any("role", req.Role),
		zap.Any("status", req.Status),
	)

	user := &model.User{
		Username: safeDerefString(req.Username),
		Email:    safeDerefString(req.Email),
		Role:     model.UserRole(safeDerefString(req.Role)),
		Status:   model.UserStatus(safeDerefString(req.Status)),
	}

	if err := h.userService.UpdateUser(id, user); err != nil {
		logger.Warn("Failed to update user",
			zap.String("user_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User updated successfully", zap.String("user_id", id))

	// 获取更新后的用户
	updatedUser, _ := h.userService.GetUser(id)

	response.Success(c, updatedUser)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	logger.Info("Delete user request", zap.String("user_id", id))

	if err := h.userService.DeleteUser(id); err != nil {
		logger.Warn("Failed to delete user",
			zap.String("user_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User deleted successfully", zap.String("user_id", id))
	response.Success(c, gin.H{"message": "User deleted successfully"})
}

// ChangePassword 修改用户密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	// 从 JWT token 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	id := userID.(string)

	logger.Info("User password change request", zap.String("user_id", id))

	if err := h.userService.ChangePassword(id, req.OldPassword, req.NewPassword); err != nil {
		logger.Warn("Failed to change password",
			zap.String("user_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User password changed successfully", zap.String("user_id", id))
	response.Success(c, gin.H{"message": "Password changed successfully"})
}

// ResetUserPassword 管理员重置用户密码
func (h *UserHandler) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request")
		return
	}

	logger.Info("Admin reset user password request", zap.String("user_id", id))

	if err := h.userService.ResetPassword(id, req.NewPassword); err != nil {
		logger.Warn("Failed to reset user password",
			zap.String("user_id", id),
			zap.String("error", err.Error()),
		)
		response.BadRequest(c, err.Error())
		return
	}

	logger.Info("User password reset successfully", zap.String("user_id", id))
	response.Success(c, gin.H{"message": "Password reset successfully"})
}

// safeDerefString 安全地解引用字符串指针
func safeDerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

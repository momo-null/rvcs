package service

import (
	"errors"
	"time"

	"rvcs/internal/model"
	"rvcs/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetUser(id string) (*model.User, error)
	GetUserList(page, pageSize int, role, status string) ([]*model.User, int64, error)
	CreateUser(user *model.User, password string) error
	UpdateUser(id string, user *model.User) error
	DeleteUser(id string) error
	ChangePassword(id, oldPassword, newPassword string) error
	ResetPassword(id, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) GetUser(id string) (*model.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *userService) GetUserList(page, pageSize int, role, status string) ([]*model.User, int64, error) {
	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取用户列表
	users, total, err := s.userRepo.List(offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	// 如果有过滤条件，需要额外过滤（简单的内存过滤，生产环境建议在 SQL 层过滤）
	if role != "" || status != "" {
		filtered := make([]*model.User, 0)
		for _, user := range users {
			if role != "" && string(user.Role) != role {
				continue
			}
			if status != "" && string(user.Status) != status {
				continue
			}
			filtered = append(filtered, user)
		}
		// 更新总数（这里简化处理，实际应该在 SQL 层统计）
		return filtered, total, nil
	}

	return users, total, nil
}

func (s *userService) CreateUser(user *model.User, password string) error {
	// 检查用户名是否已存在
	existingUser, _ := s.userRepo.FindByUsername(user.Username)
	if existingUser != nil {
		return errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	existingUser, _ = s.userRepo.FindByEmail(user.Email)
	if existingUser != nil {
		return errors.New("email already exists")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 设置密码哈希和创建时间
	user.PasswordHash = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// 创建用户
	return s.userRepo.Create(user)
}

func (s *userService) UpdateUser(id string, user *model.User) error {
	// 检查用户是否存在
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// 检查用户名是否被其他用户使用
	if user.Username != "" && user.Username != existingUser.Username {
		otherUser, _ := s.userRepo.FindByUsername(user.Username)
		if otherUser != nil && otherUser.ID != id {
			return errors.New("username already exists")
		}
	}

	// 检查邮箱是否被其他用户使用
	if user.Email != "" && user.Email != existingUser.Email {
		otherUser, _ := s.userRepo.FindByEmail(user.Email)
		if otherUser != nil && otherUser.ID != id {
			return errors.New("email already exists")
		}
	}

	// 更新字段
	if user.Username != "" {
		existingUser.Username = user.Username
	}
	if user.Email != "" {
		existingUser.Email = user.Email
	}
	if user.Role != "" {
		existingUser.Role = user.Role
	}
	if user.Status != "" {
		existingUser.Status = user.Status
	}
	if user.Avatar != nil {
		existingUser.Avatar = user.Avatar
	}
	existingUser.UpdatedAt = time.Now()

	// 更新用户
	return s.userRepo.Update(existingUser)
}

func (s *userService) DeleteUser(id string) error {
	// 检查用户是否存在
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// 删除用户
	return s.userRepo.Delete(id)
}

func (s *userService) ChangePassword(id, oldPassword, newPassword string) error {
	// 查找用户
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("incorrect old password")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

func (s *userService) ResetPassword(id, newPassword string) error {
	// 查找用户
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 更新密码
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}

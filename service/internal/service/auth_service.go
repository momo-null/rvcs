package service

import (
	"errors"
	"fmt"
	"time"

	"rvcs/internal/dto"
	"rvcs/internal/model"
	"rvcs/internal/repository"
	"rvcs/internal/security"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*model.User, error)
	Login(username, password string) (*dto.LoginResponse, error)
	RefreshToken(refreshToken string) (*dto.RefreshTokenResponse, error)
	ValidateToken(token string) (*security.Claims, error)
}

type authService struct {
	userRepo   repository.UserRepository
	jwtManager *security.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, jwtManager *security.JWTManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *authService) Register(req *dto.RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	_, err := s.userRepo.FindByUsername(req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	_, err = s.userRepo.FindByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         model.UserRoleUser,
		Status:       model.UserStatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(username, password string) (*dto.LoginResponse, error) {
	fmt.Printf("Login attempt for username: %s\n", username)
	
	// 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		fmt.Printf("User not found: %s, error: %v\n", username, err)
		return nil, errors.New("invalid username or password")
	}

	fmt.Printf("Found user: %s, ID: %s, Status: %s\n", user.Username, user.ID, user.Status)
	fmt.Printf("Stored password hash: %s\n", user.PasswordHash)
	
	// 验证密码 - 直接使用bcrypt避免security包问题
	fmt.Printf("Verifying password for user: %s\n", username)
	fmt.Printf("Input password: %s\n", password)
	fmt.Printf("Stored hash: %s\n", user.PasswordHash)
	
	// 直接使用bcrypt验证
	bcryptErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	fmt.Printf("Bcrypt error: %v\n", bcryptErr)
	isValid := bcryptErr == nil
	fmt.Printf("Direct bcrypt verification: %t\n", isValid)
	
	// 验证security包函数是否正常工作
	securityResult := security.VerifyPassword(password, user.PasswordHash)
	fmt.Printf("Security package verification: %t\n", securityResult)
	
	if !isValid {
		fmt.Printf("Password verification failed for user: %s\n", username)
		return nil, errors.New("invalid username or password")
	}

	fmt.Printf("Password verified successfully for user: %s\n", username)

	// 检查用户状态
	if user.Status != model.UserStatusActive {
		return nil, errors.New("user account is not active")
	}

	// 生成 JWT token
	accessToken, refreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(user)

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      string(user.Role),
			Avatar:    safeStringPtr(user.Avatar),
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

func (s *authService) RefreshToken(refreshToken string) (*dto.RefreshTokenResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 查找用户
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 生成新的令牌对
	accessToken, newRefreshToken, err := s.jwtManager.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) ValidateToken(token string) (*security.Claims, error) {
	return s.jwtManager.ValidateToken(token)
}

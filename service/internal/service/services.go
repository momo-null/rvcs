package service

import (
	"fmt"

	"rvcs/internal/config"
	"rvcs/internal/repository"
	"rvcs/internal/scheduler"
	"rvcs/internal/security"
	"rvcs/internal/websocket"
)

// Services 服务容器
type Services struct {
	Device           DeviceService
	Auth             AuthService
	User             UserService
	Stream           *StreamService
	Alert            *AlertService
	RegistrationCode RegistrationCodeService   // 注册码服务
	OfflineChecker   *scheduler.OfflineChecker // 离线检测器
}

// NewServices 创建服务容器
func NewServices(repos *repository.Repositories, configs *config.Config, hub *websocket.Hub) (*Services, error) {
	deviceService := NewDeviceService(repos.Device, repos.DeviceSession, repos.ActivityLog, repos.DeviceCommand, repos.RegistrationCode, hub)
	authService := NewAuthService(repos.User, security.NewJWTManager(
		configs.JWT.Secret,
		configs.JWT.AccessTokenExpiry,
		configs.JWT.RefreshTokenExpiry,
	))
	userService := NewUserService(repos.User)
	streamService := NewStreamService(repos.Stream)
	alertService := NewAlertService(repos.Alert, nil)

	// 初始化注册码服务
	registrationCodeService, err := NewRegistrationCodeService(repos.RegistrationCode, configs.RegistrationCode.Reusable)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize registration code service: %v", err)
	}

	// 初始化离线检测器
	offlineChecker := scheduler.NewOfflineChecker(repos.Device, hub, configs)

	return &Services{
		Device:           deviceService,
		Auth:             authService,
		User:             userService,
		Stream:           streamService,
		Alert:            alertService,
		RegistrationCode: registrationCodeService,
		OfflineChecker:   offlineChecker,
	}, nil
}

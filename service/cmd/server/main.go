package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"rvcs/internal/api/router"
	"rvcs/internal/config"
	"rvcs/internal/database"
	"rvcs/internal/logger"
	"rvcs/internal/repository"
	"rvcs/internal/security"
	"rvcs/internal/server"
	"rvcs/internal/service"
	"rvcs/internal/websocket"
)

func main() {
	// 创建必要的目录
	dirs := []string{"./data", "./logs", "./certs"}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Warning: failed to create directory %s: %v\n", dir, err)
		}
	}

	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		logger.InitDefaultLogger()
		panic(err.Error())
	}

	// 初始化日志
	if err := logger.InitLogger(cfg); err != nil {
		panic(err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting RVCS Server...")

	// 初始化数据库（使用文件数据库模式实现持久化）
	db, err := database.InitSQLite(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize SQLite database", zap.Error(err))
	}
	defer database.Close()

	// 执行数据库迁移
	if err := database.Migrate(); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 创建默认管理员账户
	if err := database.CreateDefaultAdminUser(db); err != nil {
		logger.Error("Failed to create default admin user", zap.Error(err))
	}

	// 初始化Redis（自动切换到内存Redis）
	redisClient := database.InitRedis(cfg)
	if redisClient != nil {
		logger.Info("Connected to Redis")
		defer database.CloseRedis()
	} else {
		logger.Warn("Failed to connect to Redis, switching to memory Redis")
		if err := database.InitMemoryRedis(cfg); err != nil {
			logger.Fatal("Failed to initialize memory Redis", zap.Error(err))
		}
	}

	_ = redisClient // 保存 redisClient 引用

	// 初始化JWT管理器
	jwtManager := security.NewJWTManager(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)

	// 设置设备Token有效期（设备长期运行，使用更长的有效期）
	jwtManager.SetDeviceTokenDuration(
		cfg.JWT.DeviceAccessTokenExpiry,
		cfg.JWT.DeviceRefreshTokenExpiry,
	)

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 初始化Repository层
	repos := repository.NewRepositories(db)

	// 加载配置
	cfg, err = config.LoadConfig("./configs")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	// 初始化Service层
	services, err := service.NewServices(repos, cfg, hub)
	if err != nil {
		logger.Fatal("Failed to initialize services", zap.Error(err))
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.Default()

	// 注册路由
	router.SetupRoutes(r, services, hub, jwtManager, cfg)

	// 创建TLS服务器配置
	tlsConfig := &server.TLSServerConfig{
		HTTPPort:     cfg.Server.HTTPPort,       // HTTP端口（重定向到HTTPS）
		HTTPSPort:    cfg.Server.HTTPSPort,      // HTTPS端口
		CertFile:     cfg.TLS.CertFile,         // SSL证书文件
		KeyFile:      cfg.TLS.KeyFile,          // SSL私钥文件
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  60 * time.Second,
		Logger:       logger.Logger,
	}

	// 创建TLS服务器
	tlsServer := server.NewTLSServer(r, tlsConfig)

	// 启动服务器
	go func() {
		if err := tlsServer.Start(); err != nil {
			logger.Fatal("Failed to start TLS server", zap.Error(err))
		}
	}()

	logger.Info("Servers started",
		zap.String("mode", cfg.Server.Mode),
		zap.Int("http_port", tlsConfig.HTTPPort),
		zap.Int("https_port", tlsConfig.HTTPSPort),
		zap.String("cert_file", tlsConfig.CertFile),
	)

	// 启动离线检测器
	services.OfflineChecker.Start()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 停止离线检测器
	services.OfflineChecker.Stop()

	if err := tlsServer.GetHTTPSServer().Shutdown(ctx); err != nil {
		logger.Error("HTTPS server forced to shutdown", zap.Error(err))
	}

	if err := tlsServer.Stop(); err != nil {
		logger.Error("Failed to stop servers", zap.Error(err))
	}

	logger.Info("Servers exited")
}
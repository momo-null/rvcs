package router

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rvcs/internal/api/handler"
	"rvcs/internal/api/middleware"
	"rvcs/internal/api/response"
	"rvcs/internal/config"
	"rvcs/internal/security"
	"rvcs/internal/service"
	"rvcs/internal/websocket"

	"github.com/gin-gonic/gin"
)

var ginHTMLTemplates *template.Template

func init() {
	ginHTMLTemplates = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .content { margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.title}}</h1>
        </div>
        <div class="content">
            {{if .message}}
                <p>{{.message}}</p>
            {{end}}
            <h2>Available Endpoints:</h2>
            <ul>
                <li><a href="/health">Health Check</a> - /health</li>
                <li><a href="/api/v1/auth/login">Login API</a> - /api/v1/auth/login</li>
                <li>WebSocket - /api/v1/ws</li>
            </ul>
            <h2>Documentation:</h2>
            <p>Please refer to the project documentation for API usage.</p>
        </div>
    </div>
</body>
</html>`))
}

// SetupRoutes configures all HTTP routes.
func SetupRoutes(r *gin.Engine, services *service.Services, hub *websocket.Hub, jwtManager *security.JWTManager, cfg *config.Config) {
	// 初始化handlers
	deviceHandler := handler.NewDeviceHandler(services.Device, services.RegistrationCode, jwtManager, &cfg.LiveKit)
	authHandler := handler.NewAuthHandler(services.Auth)
	userHandler := handler.NewUserHandler(services.User)
	streamHandler := handler.NewStreamHandler(services.Stream)
	alertHandler := handler.NewAlertHandler(services.Alert)
	registrationCodeHandler := handler.NewRegistrationCodeHandler(services.RegistrationCode)

	// 使用中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// API v1 - Web端接口 (用户认证)
	webV1 := r.Group("/api/v1/web")
	{
		// 认证相关路由（不需要认证）
		auth := webV1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		// 设备管理路由（需要用户认证）
		devices := webV1.Group("/devices")
		devices.Use(middleware.AuthMiddleware(jwtManager))
		{
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.GET("", deviceHandler.ListDevices)
			devices.POST("", deviceHandler.CreateDevice)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
			devices.POST("/:id/control", deviceHandler.ControlDevice)
			devices.GET("/:id/logs", deviceHandler.GetDeviceLogs)
			devices.GET("/:id/livekit-token", deviceHandler.GetLiveKitToken)
		}

		// 日志路由（需要用户认证）
		logs := webV1.Group("/logs")
		logs.Use(middleware.AuthMiddleware(jwtManager))
		{
			logs.GET("", deviceHandler.GetAllLogs)
		}

		monitor := webV1.Group("/monitor")
		monitor.Use(middleware.AuthMiddleware(jwtManager))
		{
			monitor.GET("/livekit-overview", deviceHandler.GetLiveKitOverview)
		}

		// 用户管理路由（需要用户认证，仅管理员）
		users := webV1.Group("/users")
		users.Use(middleware.AuthMiddleware(jwtManager))
		users.Use(middleware.AdminOnlyMiddleware())
		{
			users.GET("", userHandler.GetUserList)
			users.GET("/:id", userHandler.GetUser)
			users.POST("", userHandler.CreateUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.PUT("/:id/password", userHandler.ResetUserPassword)
		}

		// 当前用户相关路由（需要用户认证）
		me := webV1.Group("/me")
		me.Use(middleware.AuthMiddleware(jwtManager))
		{
			me.PUT("/password", userHandler.ChangePassword)
		}

		// 流管理路由（需要用户认证）
		streams := webV1.Group("/streams")
		streams.Use(middleware.AuthMiddleware(jwtManager))
		{
			streams.GET("", streamHandler.GetAllStreams)
			streams.GET("/:stream_id", streamHandler.GetStream)
			streams.DELETE("/:stream_id", streamHandler.DeleteStream)
		}

		// 告警相关路由（需要用户认证）
		alerts := webV1.Group("/alerts")
		alerts.Use(middleware.AuthMiddleware(jwtManager))
		{
			alerts.GET("", alertHandler.GetAlerts)
			alerts.GET("/stats", alertHandler.GetAlertStats)
			alerts.POST("/:alert_id/acknowledge", alertHandler.AcknowledgeAlert)
			alerts.POST("/:alert_id/resolve", alertHandler.ResolveAlert)
		}

		// 注册码管理路由（需要管理员权限）
		registrationCodes := webV1.Group("/registration-codes")
		registrationCodes.Use(middleware.AuthMiddleware(jwtManager))
		registrationCodes.Use(middleware.AdminOnlyMiddleware())
		{
			registrationCodes.POST("", registrationCodeHandler.GenerateCode)
			registrationCodes.GET("", registrationCodeHandler.ListCodes)
			registrationCodes.GET("/:id", registrationCodeHandler.GetCode)
			registrationCodes.DELETE("/:id", registrationCodeHandler.DeleteCode)
			registrationCodes.POST("/:id/revoke", registrationCodeHandler.RevokeCode)
			registrationCodes.POST("/:id/reset", registrationCodeHandler.ResetCode)
			registrationCodes.POST("/cleanup", registrationCodeHandler.CleanupExpiredCodes)
		}
	}

	// API v1 - Android端接口 (设备认证)
	deviceV1 := r.Group("/api/v1/device")
	{
		// 设备注册相关路由（公开接口，无需认证）
		deviceV1.POST("/register", deviceHandler.Register)
		deviceV1.POST("/register-with-code", deviceHandler.RegisterWithCode)

		// 设备token刷新（公开接口，无需认证）
		deviceV1.POST("/refresh", deviceHandler.RefreshDeviceToken)

		// 设备心跳路由（需要设备认证）
		heartbeat := deviceV1.Group("")
		heartbeat.Use(middleware.DeviceAuthMiddleware(jwtManager))
		{
			heartbeat.POST("/heartbeat", deviceHandler.Heartbeat)
			heartbeat.GET("/commands/pending", deviceHandler.PullPendingCommands)
			heartbeat.POST("/commands/:id/ack", deviceHandler.AckCommand)
			heartbeat.POST("/alerts", alertHandler.ReportDeviceAlert)
		}

		// 流创建和LiveKit Token路由（需要设备认证）
		streams := deviceV1.Group("/:id")
		streams.Use(middleware.DeviceAuthMiddleware(jwtManager))
		{
			streams.POST("/stream", streamHandler.CreateStream)
			streams.GET("/streams", streamHandler.GetStreams)
			streams.GET("/livekit-token", deviceHandler.GetDeviceLiveKitToken)
		}
	}

	// WebSocket路由（需要用户认证）
	r.GET("/api/v1/web/ws", func(c *gin.Context) {
		handler.HandleWebSocket(c, hub, jwtManager)
	})

	// 设备WebSocket路由（需要设备认证）
	deviceV1.GET("/ws", func(c *gin.Context) {
		handler.HandleDeviceWebSocket(c, hub, jwtManager)
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	// 静态文件服务 - 提供前端界面
	setupStaticFiles(r)
}

// setupStaticFiles 配置静态文件服务
func setupStaticFiles(r *gin.Engine) {
	// 检查web目录是否存在
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// 如果本地web目录不存在，尝试上级目录
		webDir = "../web/dist"
		if _, err := os.Stat(webDir); os.IsNotExist(err) {
			// 如果都不存在，创建基本的欢迎页面
			r.GET("/", func(c *gin.Context) {
				c.HTML(http.StatusOK, "welcome.html", gin.H{
					"title": "RVCS - Remote Vision & Control System",
				})
			})
			return
		}
	}

	// 服务静态文件
	r.StaticFS("/static", http.Dir(filepath.Join(webDir, "assets")))
	// 为前端资源提供/assets路径
	r.StaticFS("/assets", http.Dir(filepath.Join(webDir, "assets")))

	// 主页路由 - 提供index.html
	r.GET("/", func(c *gin.Context) {
		indexPath := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}
		c.HTML(http.StatusOK, "fallback.html", gin.H{
			"title":   "RVCS - Remote Vision & Control System",
			"message": "Frontend files not found. Please build the web frontend first.",
		})
	})

	// SPA路由支持 - 所有未匹配的路由都返回index.html
	r.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// 其他请求返回前端应用
		indexPath := filepath.Join(webDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			c.File(indexPath)
			return
		}
		c.HTML(http.StatusOK, "fallback.html", gin.H{
			"title":   "RVCS - Remote Vision & Control System",
			"message": "Frontend files not found. Please build the web frontend first.",
		})
	})
}

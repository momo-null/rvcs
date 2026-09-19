package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server           ServerConfig           `mapstructure:"server"`
	Database         DatabaseConfig         `mapstructure:"database"`
	Redis            RedisConfig            `mapstructure:"redis"`
	JWT              JWTConfig              `mapstructure:"jwt"`
	WebSocket        WebSocketConfig        `mapstructure:"websocket"`
	RateLimit        RateLimitConfig        `mapstructure:"rate_limit"`
	Log              LogConfig              `mapstructure:"log"`
	Stream           StreamConfig           `mapstructure:"stream"`
	TLS              TLSConfig              `mapstructure:"tls"`
	RegistrationCode RegistrationCodeConfig `mapstructure:"registration_code"`
	LiveKit          LiveKitConfig          `mapstructure:"livekit"`
}

type ServerConfig struct {
	HTTPPort     int           `mapstructure:"http_port"`
	HTTPSPort    int           `mapstructure:"https_port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Mode            string        `mapstructure:"mode"` // 数据库模式: sqlite-file, sqlite-memory, mysql
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	// SQLite特定配置
	SQLiteDataDir  string `mapstructure:"sqlite_data_dir"`
	SQLiteFilename string `mapstructure:"sqlite_filename"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret                   string        `mapstructure:"secret"`
	AccessTokenExpiry        time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry       time.Duration `mapstructure:"refresh_token_expiry"`
	DeviceAccessTokenExpiry  time.Duration `mapstructure:"device_access_token_expiry"`
	DeviceRefreshTokenExpiry time.Duration `mapstructure:"device_refresh_token_expiry"`
}

type WebSocketConfig struct {
	ReadBufferSize       int           `mapstructure:"read_buffer_size"`
	WriteBufferSize      int           `mapstructure:"write_buffer_size"`
	PingInterval         time.Duration `mapstructure:"ping_interval"`
	PongWait             time.Duration `mapstructure:"pong_wait"`
	WriteWait            time.Duration `mapstructure:"write_wait"`
	MaxMessageSize       int64         `mapstructure:"max_message_size"`
	OfflineCheckInterval time.Duration `mapstructure:"offline_check_interval"` // 设备离线检测间隔
	OfflineTimeout       time.Duration `mapstructure:"offline_timeout"`        // 设备离线超时时间
}

type RateLimitConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	Window      time.Duration `mapstructure:"window"`
	MaxRequests int           `mapstructure:"max_requests"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type StreamConfig struct {
	RTSPPort           int           `mapstructure:"rtsp_port"`
	RTMPPort           int           `mapstructure:"rtmp_port"`
	HLSPort            int           `mapstructure:"hls_port"`
	HLSSegmentDuration time.Duration `mapstructure:"hls_segment_duration"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type RegistrationCodeConfig struct {
	Reusable bool `mapstructure:"reusable"` // 注册码是否可重复使用
}

type LiveKitConfig struct {
	ServerURL       string        `mapstructure:"server_url"`
	PublicServerURL string        `mapstructure:"public_server_url"`
	APIKey          string        `mapstructure:"api_key"`
	APISecret       string        `mapstructure:"api_secret"`
	RoomPrefix      string        `mapstructure:"room_name_prefix"`
	TokenExpiry     time.Duration `mapstructure:"token_expiry"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
// 支持从 ../configs 和 ./configs 加载，如果目录不存在不报错
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 添加多个配置路径，按优先级顺序
	// 当前目录的 configs
	relativePath := filepath.Join(".", "configs")
	absPath, err := filepath.Abs(relativePath)
	if err == nil && dirExists(absPath) {
		v.AddConfigPath(relativePath)
	}

	// 父目录的 configs
	parentPath := filepath.Join("..", "configs")
	absParentPath, err := filepath.Abs(parentPath)
	if err == nil && dirExists(absParentPath) {
		v.AddConfigPath(parentPath)
	}

	// 传入的配置路径
	if configPath != "" {
		absConfigPath, err := filepath.Abs(configPath)
		if err == nil && dirExists(absConfigPath) {
			v.AddConfigPath(configPath)
		}
	}

	// 设置环境变量
	v.SetEnvPrefix("CAMERA_MONITOR")
	v.AutomaticEnv()

	// 尝试从第一个路径读取
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，返回默认配置而不是报错
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return getDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = config
	return config, nil
}

// dirExists 检查目录是否存在
func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getDefaultConfig 返回默认配置
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTPPort:     8080,
			HTTPSPort:    8443,
			Mode:         "debug",
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Database: DatabaseConfig{
			Mode:            "sqlite-file",
			Host:            "localhost",
			Port:            3306,
			User:            "root",
			Password:        "password",
			DBName:          "rvcs",
			SSLMode:         "disable",
			MaxIdleConns:    10,
			MaxOpenConns:    100,
			ConnMaxLifetime: time.Hour,
			SQLiteDataDir:   "./data",
			SQLiteFilename:  "rvcs.db",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		JWT: JWTConfig{
			Secret:             "your-secret-key-change-this",
			AccessTokenExpiry:  time.Hour,
			RefreshTokenExpiry: 24 * time.Hour,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:       1024,
			WriteBufferSize:      1024,
			PingInterval:         30 * time.Second,
			PongWait:             60 * time.Second,
			WriteWait:            10 * time.Second,
			MaxMessageSize:       512,
			OfflineCheckInterval: 60 * time.Second,  // 每60秒检测一次
			OfflineTimeout:       180 * time.Second, // 180秒未心跳视为离线，降低熄屏场景误判
		},
		RateLimit: RateLimitConfig{
			Enabled:     true,
			Window:      time.Minute,
			MaxRequests: 100,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
		Stream: StreamConfig{
			RTSPPort:           554,
			RTMPPort:           1935,
			HLSPort:            8081,
			HLSSegmentDuration: 10 * time.Second,
		},
		TLS: TLSConfig{
			Enabled:  false,
			CertFile: "",
			KeyFile:  "",
		},
		RegistrationCode: RegistrationCodeConfig{
			Reusable: false, // 默认不可重复使用
		},
	}
}

// GetDSN 获取数据库连接字符串 (MySQL)
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName,
	)
}

// GetRedisAddr 获取Redis地址
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

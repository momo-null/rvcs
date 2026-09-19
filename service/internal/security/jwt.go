package security

import (
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id,omitempty"`
	Role     string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// JWTManager JWT管理器
type JWTManager struct {
	secretKey                 string
	accessTokenDuration       time.Duration
	refreshTokenDuration      time.Duration
	deviceAccessTokenDuration  time.Duration
	deviceRefreshTokenDuration time.Duration
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, accessDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:                 secretKey,
		accessTokenDuration:       accessDuration,
		refreshTokenDuration:      refreshDuration,
		deviceAccessTokenDuration:  accessDuration, // 默认使用相同配置
		deviceRefreshTokenDuration: refreshDuration, // 默认使用相同配置
	}
}

// SetDeviceTokenDuration 设置设备Token有效期
func (m *JWTManager) SetDeviceTokenDuration(accessDuration, refreshDuration time.Duration) {
	m.deviceAccessTokenDuration = accessDuration
	m.deviceRefreshTokenDuration = refreshDuration
}

// GenerateTokenPair 生成Token对
func (m *JWTManager) GenerateTokenPair(userID, role string) (string, string, error) {
	// 生成Access Token
	accessToken, err := m.generateToken(userID, role, m.accessTokenDuration)
	if err != nil {
		return "", "", err
	}

	// 生成Refresh Token
	refreshToken, err := m.generateToken(userID, role, m.refreshTokenDuration)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateDeviceToken 生成设备Access Token
func (m *JWTManager) GenerateDeviceToken(deviceID string) (string, error) {
	return m.generateDeviceToken(deviceID, m.deviceAccessTokenDuration)
}

// GenerateDeviceRefreshToken 生成设备Refresh Token
func (m *JWTManager) GenerateDeviceRefreshToken(deviceID string) (string, error) {
	return m.generateDeviceToken(deviceID, m.deviceRefreshTokenDuration)
}

// GenerateDeviceTokenPair 生成设备Token对
func (m *JWTManager) GenerateDeviceTokenPair(deviceID string) (string, string, error) {
	// 生成Access Token
	accessToken, err := m.GenerateDeviceToken(deviceID)
	if err != nil {
		return "", "", err
	}

	// 生成Refresh Token
	refreshToken, err := m.GenerateDeviceRefreshToken(deviceID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateDeviceToken 生成设备Token
func (m *JWTManager) generateDeviceToken(deviceID string, duration time.Duration) (string, error) {
	claims := Claims{
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// generateToken 生成Token
func (m *JWTManager) generateToken(userID, role string, duration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// ValidateToken 验证Token
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	log.Printf("[JWT VALIDATE] 开始验证token: %.50s...", tokenString)

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[JWT VALIDATE] 无效的签名方法: %v", token.Method)
			return nil, ErrInvalidToken
		}
		return []byte(m.secretKey), nil
	})

	if err != nil {
		log.Printf("[JWT VALIDATE] Token解析错误: %v", err)
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Printf("[JWT VALIDATE] Token已过期")
			return nil, ErrExpiredToken
		}
		log.Printf("[JWT VALIDATE] 无效的Token")
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Printf("[JWT VALIDATE] Token声明无效或token无效")
		return nil, ErrInvalidToken
	}

	log.Printf("[JWT VALIDATE] Token验证成功 - UserID: %s, DeviceID: %s, Role: %s",
		claims.UserID, claims.DeviceID, claims.Role)
	return claims, nil
}

// RefreshToken 刷新Token
func (m *JWTManager) RefreshToken(refreshTokenString string) (string, string, error) {
	claims, err := m.ValidateToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	return m.GenerateTokenPair(claims.UserID, claims.Role)
}

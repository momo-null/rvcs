package handler

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	gorilla_websocket "github.com/gorilla/websocket"
	"go.uber.org/zap"

	"rvcs/internal/logger"
	"rvcs/internal/security"
	ws "rvcs/internal/websocket"
)

var upgrader = gorilla_websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return isWebSocketOriginAllowed(r)
	},
}

// HandleDeviceWebSocket handles device websocket connection.
func HandleDeviceWebSocket(c *gin.Context, hub *ws.Hub, jwtManager *security.JWTManager) {
	authHeader := c.GetHeader("Authorization")
	token := ""

	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token = parts[1]
		}
	}

	// Keep query-token fallback for compatibility with existing clients.
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		logger.Warn("Device WebSocket connection attempt without token", zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		logger.Warn("Device WebSocket connection attempt with invalid token",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid device token"})
		return
	}

	deviceID := claims.DeviceID
	if deviceID == "" {
		logger.Warn("Device WebSocket connection with invalid device ID",
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid device ID"})
		return
	}

	logger.Info("Device WebSocket connection established",
		zap.String("device_id", deviceID),
		zap.String("client_ip", c.ClientIP()),
	)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Device WebSocket upgrade error",
			zap.String("device_id", deviceID),
			zap.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	client := ws.NewClient(hub, deviceID, conn)
	client.DeviceID = deviceID

	hub.Register(client)

	logger.Debug("Device WebSocket client registered to hub",
		zap.String("device_id", deviceID),
	)

	go client.WritePump()
	go client.ReadPump()
}

// HandleWebSocket handles web websocket connection.
func HandleWebSocket(c *gin.Context, hub *ws.Hub, jwtManager *security.JWTManager) {
	authHeader := c.GetHeader("Authorization")
	token := ""

	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token = parts[1]
		}
	}

	// Browser WebSocket cannot set Authorization headers.
	if token == "" {
		token = c.Query("token")
	}

	if token == "" {
		logger.Warn("WebSocket connection attempt without token", zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		logger.Warn("WebSocket connection attempt with invalid token",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	deviceID := c.Query("device_id")

	logger.Info("WebSocket connection established",
		zap.String("device_id", deviceID),
		zap.String("user_id", claims.UserID),
		zap.String("client_ip", c.ClientIP()),
	)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket upgrade error",
			zap.String("device_id", deviceID),
			zap.String("error", err.Error()),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "WebSocket upgrade failed"})
		return
	}

	clientID := deviceID
	if clientID == "" {
		clientID = claims.UserID
	}

	client := ws.NewClient(hub, clientID, conn)
	client.UserID = claims.UserID
	client.DeviceID = deviceID

	hub.Register(client)

	logger.Debug("WebSocket client registered to hub",
		zap.String("client_id", clientID),
		zap.String("device_id", deviceID),
	)

	go client.WritePump()
	go client.ReadPump()
}

func isWebSocketOriginAllowed(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	originURL, err := url.Parse(origin)
	if err != nil || originURL.Host == "" {
		return false
	}
	originHost := strings.ToLower(originURL.Host)
	reqHost := strings.ToLower(getRequestHost(r))
	if originHost == reqHost {
		return true
	}

	allowedOrigins := parseWSAllowedOrigins(os.Getenv("WS_ALLOWED_ORIGINS"))
	if _, ok := allowedOrigins[originHost]; ok {
		return true
	}
	if _, ok := allowedOrigins[strings.ToLower(origin)]; ok {
		return true
	}
	return false
}

func parseWSAllowedOrigins(raw string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, part := range strings.Split(raw, ",") {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		out[strings.ToLower(item)] = struct{}{}
	}
	return out
}

func getRequestHost(r *http.Request) string {
	if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
		return strings.Split(h, ",")[0]
	}
	return r.Host
}

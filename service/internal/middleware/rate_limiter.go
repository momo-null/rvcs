package middleware

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"rvcs/internal/logger"
)

// RateLimiter tracks request counts by client id.
type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	window  time.Duration
	limit   int
}

type ClientInfo struct {
	requests    []time.Time
	windowStart time.Time
}

func NewRateLimiter(window time.Duration, limit int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		window:  window,
		limit:   limit,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, client := range rl.clients {
			if now.Sub(client.windowStart) > rl.window {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(clientID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]
	if !exists || now.Sub(client.windowStart) > rl.window {
		rl.clients[clientID] = &ClientInfo{requests: []time.Time{now}, windowStart: now}
		return true
	}

	if len(client.requests) >= rl.limit {
		return false
	}

	client.requests = append(client.requests, now)
	return true
}

// RateLimitMiddleware is disabled by default.
// Set RVCS_ENABLE_RATE_LIMIT=1 to re-enable.
func RateLimitMiddleware(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("RVCS_ENABLE_RATE_LIMIT") != "1" {
			c.Next()
			return
		}

		clientID := c.GetHeader("X-Client-ID")
		if clientID == "" {
			if deviceID, exists := c.Get("device_id"); exists {
				clientID = deviceID.(string)
			} else if userID, exists := c.Get("user_id"); exists {
				clientID = userID.(string)
			} else {
				clientID = c.ClientIP()
			}
		}

		if !rl.Allow(clientID) {
			logger.Warn("Rate limit exceeded",
				zap.String("client_id", clientID),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "Too many requests",
				"retry_after": int(rl.window.Seconds()),
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		remaining := rl.getRemainingRequests(clientID)
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.window).Unix()))
		c.Next()
	}
}

func (rl *RateLimiter) getRemainingRequests(clientID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	client, exists := rl.clients[clientID]
	if !exists {
		return rl.limit
	}

	now := time.Now()
	validRequests := make([]time.Time, 0, len(client.requests))
	for _, req := range client.requests {
		if now.Sub(req) <= rl.window {
			validRequests = append(validRequests, req)
		}
	}

	return rl.limit - len(validRequests)
}

type DeviceRateLimiter struct{ *RateLimiter }

type UserRateLimiter struct{ *RateLimiter }

type GlobalRateLimiter struct{ *RateLimiter }

func NewDeviceRateLimiter() *DeviceRateLimiter {
	return &DeviceRateLimiter{RateLimiter: NewRateLimiter(15*time.Minute, 100)}
}

func NewUserRateLimiter() *UserRateLimiter {
	return &UserRateLimiter{RateLimiter: NewRateLimiter(time.Minute, 60)}
}

func NewGlobalRateLimiter() *GlobalRateLimiter {
	return &GlobalRateLimiter{RateLimiter: NewRateLimiter(time.Second, 100)}
}

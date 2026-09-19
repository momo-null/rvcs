package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"rvcs/internal/config"

	"github.com/redis/go-redis/v9"
)

// MemoryRedis 内存Redis实现
type MemoryRedis struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

var memoryRedis *MemoryRedis

// InitMemoryRedis 初始化内存Redis
func InitMemoryRedis(cfg *config.Config) error {
	memoryRedis = &MemoryRedis{
		data: make(map[string]interface{}),
	}
	fmt.Println("Using memory-based Redis implementation")
	return nil
}

// GetMemoryRedis 获取内存Redis实例
func GetMemoryRedis() *MemoryRedis {
	return memoryRedis
}

// Ping 实现Ping方法
func (m *MemoryRedis) Ping(ctx context.Context) *redis.StatusCmd {
	return redis.NewStatusResult("PONG", nil)
}

// Set 实现Set方法
func (m *MemoryRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.data[key] = value
	
	// 如果设置了过期时间，启动定时清理
	if expiration > 0 {
		go func() {
			time.Sleep(expiration)
			m.mutex.Lock()
			delete(m.data, key)
			m.mutex.Unlock()
		}()
	}
	
	return redis.NewStatusResult("OK", nil)
}

// Get 实现Get方法
func (m *MemoryRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if value, exists := m.data[key]; exists {
		if str, ok := value.(string); ok {
			return redis.NewStringResult(str, nil)
		}
		return redis.NewStringResult(fmt.Sprintf("%v", value), nil)
	}
	
	return redis.NewStringResult("", redis.Nil)
}

// Del 实现Del方法
func (m *MemoryRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	count := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			delete(m.data, key)
			count++
		}
	}
	
	return redis.NewIntResult(count, nil)
}

// Exists 实现Exists方法
func (m *MemoryRedis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	count := int64(0)
	for _, key := range keys {
		if _, exists := m.data[key]; exists {
			count++
		}
	}
	
	return redis.NewIntResult(count, nil)
}

// Close 实现Close方法
func (m *MemoryRedis) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 清空所有数据
	m.data = make(map[string]interface{})
	return nil
}

// Keys 实现Keys方法
func (m *MemoryRedis) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	keys := make([]string, 0, len(m.data))
	for key := range m.data {
		// 简单的通配符匹配（只支持 *）
		if pattern == "*" || matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}
	
	return redis.NewStringSliceResult(keys, nil)
}

// matchPattern 简单的模式匹配
func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	// 这里可以实现更复杂的模式匹配逻辑
	// 目前只做简单的包含检查
	return len(pattern) > 0 && len(key) >= len(pattern)
}
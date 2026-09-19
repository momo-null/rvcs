package middleware

import (
	"bytes"
	"io"
	"strings"

	"rvcs/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger 日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// 记录请求入参（仅对 POST/PUT/PATCH 方法）
		if method == "POST" || method == "PUT" || method == "PATCH" {
			if c.Request.Body != nil {
				// 读取并记录请求体
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					// 记录请求体（脱敏处理）
					sanitizedBody := sanitizeSensitiveData(string(bodyBytes))
					logger.Info("HTTP Request Body",
						zap.String("method", method),
						zap.String("path", path),
						zap.String("body", sanitizedBody),
						zap.String("client_ip", c.ClientIP()),
					)
					// 重新设置请求体，供后续处理使用
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		c.Next()

		status := c.Writer.Status()
		logger.Info("HTTP Request Complete",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// sanitizeSensitiveData 脱敏敏感数据
func sanitizeSensitiveData(body string) string {
	// 简单的脱敏处理：替换常见的敏感字段
	body = strings.ReplaceAll(body, `"password":"`, `"password":"***`)
	body = strings.ReplaceAll(body, `"old_password":"`, `"old_password":"***`)
	body = strings.ReplaceAll(body, `"new_password":"`, `"new_password":"***`)
	body = strings.ReplaceAll(body, `"token":"`, `"token":"***`)
	body = strings.ReplaceAll(body, `"refresh_token":"`, `"refresh_token":"***`)
	body = strings.ReplaceAll(body, `"access_token":"`, `"access_token":"***`)

	// 如果请求体太长，截断
	if len(body) > 500 {
		body = body[:500] + "..."
	}

	return body
}

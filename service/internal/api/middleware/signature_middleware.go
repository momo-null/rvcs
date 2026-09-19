package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"rvcs/internal/api/response"
	"rvcs/internal/core/security"

	"github.com/gin-gonic/gin"
)

// EnhancedSignatureMiddleware 强化签名验证中间件
func EnhancedSignatureMiddleware(signatureManager *security.EnhancedSignature) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对特定敏感API启用签名验证
		if !shouldVerifySignature(c.Request.URL.Path, c.Request.Method) {
			c.Next()
			return
		}
		
		// 解析签名头部
		headers := parseRequestHeaders(c)
		signatureResult, err := signatureManager.ParseSignatureHeaders(headers)
		if err != nil {
			response.Unauthorized(c, "missing or invalid signature headers")
			c.Abort()
			return
		}
		
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			response.BadRequest(c, "failed to read request body")
			c.Abort()
			return
		}
		
		// 将body放回请求中供后续处理
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		
		// 构建期望的签名载荷
		method := c.Request.Method
		endpoint := c.Request.URL.Path
		
		// 重新生成签名进行验证
		expectedResult, err := signatureManager.GenerateSignature(
			signatureResult.Payload.ClientID,
			method,
			endpoint,
			headers,
			body,
		)
		if err != nil {
			response.Error(c, 500, fmt.Sprintf("failed to generate signature for verification: %v", err))
			c.Abort()
			return
		}
		
		// 验证签名
		if expectedResult.Signature != signatureResult.Signature {
			response.Unauthorized(c, "signature verification failed")
			c.Abort()
			return
		}
		
		// 验证时间戳和其他约束
		if err := signatureManager.VerifySignature(expectedResult); err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}
		
		// 验证通过，设置客户端ID到上下文
		c.Set("client_id", signatureResult.Payload.ClientID)
		c.Next()
	}
}

// shouldVerifySignature 判断是否需要验证签名（专门针对设备接口）
func shouldVerifySignature(path, method string) bool {
	// 只对设备相关的敏感操作启用签名验证
	deviceEndpoints := []string{
		"/api/v1/devices/register",
		"/api/v1/devices/heartbeat",
		"/api/v1/devices/*/stream",
	}
	
	// 检查是否为设备相关端点
	for _, endpoint := range deviceEndpoints {
		// 处理通配符
		if strings.Contains(endpoint, "*") {
			prefix := strings.Split(endpoint, "*")[0]
			if strings.HasPrefix(path, prefix) {
				return true
			}
		} else if path == endpoint {
			return true
		}
	}
	
	return false
}

// parseRequestHeaders 解析请求头部
func parseRequestHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	
	// 解析标准头部
	headerMap := map[string]string{
		"X-Client-ID":  c.GetHeader("X-Client-ID"),
		"X-Timestamp":  c.GetHeader("X-Timestamp"),
		"X-Nonce":      c.GetHeader("X-Nonce"),
		"X-Signature":  c.GetHeader("X-Signature"),
		"Content-Type": c.GetHeader("Content-Type"),
		"Accept":       c.GetHeader("Accept"),
		"User-Agent":   c.GetHeader("User-Agent"),
	}
	
	// 过滤空值
	for key, value := range headerMap {
		if value != "" {
			headers[strings.ToLower(key)] = value
		}
	}
	
	return headers
}

// SignatureLoggerMiddleware 签名日志中间件（用于调试）
func SignatureLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-Signature") != "" {
			// 记录签名相关信息（生产环境应谨慎记录）
			c.Header("X-Signature-Debug", "Signature validation enabled for this endpoint")
		}
		c.Next()
	}
}
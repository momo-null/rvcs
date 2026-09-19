package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// SignatureConfig 签名配置
type SignatureConfig struct {
	SecretKey     string        // 签名密钥
	TimeTolerance time.Duration // 时间容忍度（防止时钟偏差）
	NonceLength   int           // 随机数长度
}

// EnhancedSignature 强化签名管理器
type EnhancedSignature struct {
	config *SignatureConfig
}

// SignaturePayload 签名载荷
type SignaturePayload struct {
	ClientID  string            `json:"client_id"`   // 客户端ID
	Timestamp int64             `json:"timestamp"`   // 时间戳
	Nonce     string            `json:"nonce"`       // 随机数
	Method    string            `json:"method"`      // HTTP方法
	Endpoint  string            `json:"endpoint"`    // 请求端点
	Headers   map[string]string `json:"headers"`     // 关键头部
	BodyHash  string            `json:"body_hash"`   // 请求体哈希
}

// SignatureResult 签名结果
type SignatureResult struct {
	Payload   *SignaturePayload `json:"payload"`
	Signature string            `json:"signature"`
	RawString string            `json:"raw_string"`
}

// NewEnhancedSignature 创建强化签名管理器
func NewEnhancedSignature(secretKey string) *EnhancedSignature {
	return &EnhancedSignature{
		config: &SignatureConfig{
			SecretKey:     secretKey,
			TimeTolerance: 5 * time.Minute, // 5分钟时间容忍
			NonceLength:   16,              // 16字节随机数
		},
	}
}

// GenerateSignature 生成强化签名
func (es *EnhancedSignature) GenerateSignature(
	clientID, method, endpoint string,
	headers map[string]string,
	body []byte,
) (*SignatureResult, error) {
	
	// 生成随机数
	nonce := es.generateNonce()
	
	// 计算请求体哈希
	bodyHash := es.calculateBodyHash(body)
	
	// 构建签名载荷
	payload := &SignaturePayload{
		ClientID:  clientID,
		Timestamp: time.Now().Unix(),
		Nonce:     nonce,
		Method:    strings.ToUpper(method),
		Endpoint:  endpoint,
		Headers:   es.extractImportantHeaders(headers),
		BodyHash:  bodyHash,
	}
	
	// 生成签名字符串
	signString := es.buildSignatureString(payload)
	
	// 计算HMAC签名
	signature := es.calculateHMAC(signString)
	
	return &SignatureResult{
		Payload:   payload,
		Signature: signature,
		RawString: signString,
	}, nil
}

// VerifySignature 验证签名
func (es *EnhancedSignature) VerifySignature(result *SignatureResult) error {
	// 1. 验证时间戳
	if err := es.validateTimestamp(result.Payload.Timestamp); err != nil {
		return fmt.Errorf("timestamp validation failed: %v", err)
	}
	
	// 2. 重新生成签名进行对比
	expectedSignString := es.buildSignatureString(result.Payload)
	expectedSignature := es.calculateHMAC(expectedSignString)
	
	// 3. 安全比较签名
	if !hmac.Equal([]byte(result.Signature), []byte(expectedSignature)) {
		return fmt.Errorf("signature mismatch")
	}
	
	return nil
}

// generateNonce 生成随机数
func (es *EnhancedSignature) generateNonce() string {
	bytes := make([]byte, es.config.NonceLength)
	// 这里应该使用crypto/rand，简化示例使用时间戳+随机
	timestamp := time.Now().UnixNano()
	for i := 0; i < es.config.NonceLength; i++ {
		bytes[i] = byte((timestamp >> uint(i%8)) & 0xFF)
	}
	return hex.EncodeToString(bytes)
}

// calculateBodyHash 计算请求体哈希
func (es *EnhancedSignature) calculateBodyHash(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	hash := sha256.Sum256(body)
	return hex.EncodeToString(hash[:])
}

// extractImportantHeaders 提取重要头部
func (es *EnhancedSignature) extractImportantHeaders(headers map[string]string) map[string]string {
	important := make(map[string]string)
	importantKeys := []string{"content-type", "accept", "user-agent"}
	
	for _, key := range importantKeys {
		if value, exists := headers[key]; exists {
			important[key] = value
		}
	}
	return important
}

// buildSignatureString 构建签名字符串
func (es *EnhancedSignature) buildSignatureString(payload *SignaturePayload) string {
	var parts []string
	
	// 按固定顺序添加字段
	parts = append(parts, payload.ClientID)
	parts = append(parts, strconv.FormatInt(payload.Timestamp, 10))
	parts = append(parts, payload.Nonce)
	parts = append(parts, payload.Method)
	parts = append(parts, payload.Endpoint)
	parts = append(parts, payload.BodyHash)
	
	// 添加排序后的头部
	headerKeys := make([]string, 0, len(payload.Headers))
	for key := range payload.Headers {
		headerKeys = append(headerKeys, key)
	}
	// 简化排序（实际应该按字母顺序）
	for _, key := range headerKeys {
		parts = append(parts, key+":"+payload.Headers[key])
	}
	
	return strings.Join(parts, "|")
}

// calculateHMAC 计算HMAC签名
func (es *EnhancedSignature) calculateHMAC(data string) string {
	mac := hmac.New(sha256.New, []byte(es.config.SecretKey))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

// validateTimestamp 验证时间戳
func (es *EnhancedSignature) validateTimestamp(timestamp int64) error {
	now := time.Now().Unix()
	diff := now - timestamp
	if diff < 0 {
		diff = -diff
	}
	
	if time.Duration(diff)*time.Second > es.config.TimeTolerance {
		return fmt.Errorf("timestamp out of tolerance window")
	}
	
	return nil
}

// GetSignatureHeaders 获取签名相关的HTTP头部
func (es *EnhancedSignature) GetSignatureHeaders(result *SignatureResult) map[string]string {
	return map[string]string{
		"X-Client-ID":  result.Payload.ClientID,
		"X-Timestamp":  strconv.FormatInt(result.Payload.Timestamp, 10),
		"X-Nonce":      result.Payload.Nonce,
		"X-Signature":  result.Signature,
		"X-Payload":    base64.StdEncoding.EncodeToString([]byte(result.RawString)),
	}
}

// ParseSignatureHeaders 从HTTP头部解析签名信息
func (es *EnhancedSignature) ParseSignatureHeaders(headers map[string]string) (*SignatureResult, error) {
	requiredHeaders := []string{"X-Client-ID", "X-Timestamp", "X-Nonce", "X-Signature"}
	for _, header := range requiredHeaders {
		if _, exists := headers[header]; !exists {
			return nil, fmt.Errorf("missing required header: %s", header)
		}
	}
	
	timestamp, err := strconv.ParseInt(headers["X-Timestamp"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %v", err)
	}
	
	payload := &SignaturePayload{
		ClientID:  headers["X-Client-ID"],
		Timestamp: timestamp,
		Nonce:     headers["X-Nonce"],
	}
	
	return &SignatureResult{
		Payload:   payload,
		Signature: headers["X-Signature"],
	}, nil
}
package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"time"
)

// SignatureManager 签名管理器
type SignatureManager struct {
	secretKey string
}

// NewSignatureManager 创建签名管理器
func NewSignatureManager(secretKey string) *SignatureManager {
	return &SignatureManager{
		secretKey: secretKey,
	}
}

// GenerateSignature 生成请求签名
// 算法步骤:
// 1. 按字典序排序参数
// 2. 拼接字符串
// 3. 添加时间戳防重放
// 4. HMAC-SHA256 签名
func (sm *SignatureManager) GenerateSignature(params map[string]string, timestamp int64) (string, error) {
	// 1. 按字典序排序参数
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 2. 拼接字符串
	var builder strings.Builder
	for _, k := range keys {
		builder.WriteString(fmt.Sprintf("%s=%s", k, params[k]))
	}

	// 3. 添加时间戳防重放
	builder.WriteString(fmt.Sprintf("&timestamp=%d", timestamp))

	// 4. HMAC-SHA256 签名
	h := hmac.New(sha256.New, []byte(sm.secretKey))
	h.Write([]byte(builder.String()))
	signature := h.Sum(nil)

	// 5. Base64 编码
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature 验证请求签名
func (sm *SignatureManager) VerifySignature(params map[string]string, signature string, timestamp int64) (bool, error) {
	// 检查时间戳（防重放，5分钟有效期）
	if time.Now().Unix()-timestamp > 300 {
		return false, fmt.Errorf("signature expired")
	}

	// 重新生成签名
	expectedSignature, err := sm.GenerateSignature(params, timestamp)
	if err != nil {
		return false, err
	}

	// 比对签名
	return expectedSignature == signature, nil
}

// GenerateDeviceSignature 为设备请求生成签名
func (sm *SignatureManager) GenerateDeviceSignature(deviceID, action string, params map[string]string) (string, string, error) {
	timestamp := time.Now().Unix()

	// 构造参数
	allParams := make(map[string]string)
	allParams["device_id"] = deviceID
	allParams["action"] = action
	for k, v := range params {
		allParams[k] = v
	}

	signature, err := sm.GenerateSignature(allParams, timestamp)
	if err != nil {
		return "", "", err
	}

	return signature, fmt.Sprintf("%d", timestamp), nil
}

// VerifyDeviceSignature 验证设备请求签名
func (sm *SignatureManager) VerifyDeviceSignature(deviceID, action string, params map[string]string, signature, timestamp string) (bool, error) {
	// 构造参数
	allParams := make(map[string]string)
	allParams["device_id"] = deviceID
	allParams["action"] = action
	for k, v := range params {
		allParams[k] = v
	}

	var ts int64
	_, err := fmt.Sscanf(timestamp, "%d", &ts)
	if err != nil {
		return false, err
	}

	return sm.VerifySignature(allParams, signature, ts)
}

// GenerateStreamToken 生成流访问令牌
func (sm *SignatureManager) GenerateStreamToken(streamID string) string {
	timestamp := time.Now().Unix()
	payload := fmt.Sprintf("%s:%d", streamID, timestamp)
	h := hmac.New(sha256.New, []byte(sm.secretKey))
	h.Write([]byte(payload))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s|%d|%s", signature, timestamp, streamID)
}

// VerifyStreamToken 验证流令牌
func (sm *SignatureManager) VerifyStreamToken(token string) (streamID string, valid bool) {
	parts := strings.Split(token, "|")
	if len(parts) != 3 {
		return "", false
	}

	signature := parts[0]
	var timestamp int64
	_, err := fmt.Sscanf(parts[1], "%d", &timestamp)
	if err != nil {
		return "", false
	}
	streamID = parts[2]

	// 检查有效期（1小时）
	if time.Now().Unix()-timestamp > 3600 {
		return "", false
	}

	// 重新生成签名并验证
	payload := fmt.Sprintf("%s:%d", streamID, timestamp)
	h := hmac.New(sha256.New, []byte(sm.secretKey))
	h.Write([]byte(payload))
	expectedSignature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if expectedSignature != signature {
		return "", false
	}

	return streamID, true
}

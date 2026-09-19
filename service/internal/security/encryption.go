package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionManager 加密管理器
type EncryptionManager struct {
	key []byte
}

// NewEncryptionManager 创建加密管理器
// 注意：密钥必须是 16、24 或 32 字节长（对应 AES-128、AES-192、AES-256）
func NewEncryptionManager(secret string) *EncryptionManager {
	// 使用 SHA256 生成 32 字节密钥
	hash := sha256.Sum256([]byte(secret))
	return &EncryptionManager{
		key: hash[:], // 使用前 32 字节
	}
}

// Encrypt AES-256-CBC 加密
func (em *EncryptionManager) Encrypt(plaintext string) (string, error) {
	// 将明文转换为字节
	data := []byte(plaintext)

	// 创建加密块
	block, err := aes.NewCipher(em.key)
	if err != nil {
		return "", err
	}

	// 创建 GCM 模式（更安全，无需 IV）
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt AES-256-CBC 解密
func (em *EncryptionManager) Decrypt(ciphertext string) (string, error) {
	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建加密块
	block, err := aes.NewCipher(em.key)
	if err != nil {
		return "", err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查密文长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 提取 nonce 和密文
	nonce, ciphertext_bytes := data[:nonceSize], data[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext_bytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptToken 加密令牌
func (em *EncryptionManager) EncryptToken(token string) (string, error) {
	return em.Encrypt(token)
}

// DecryptToken 解密令牌
func (em *EncryptionManager) DecryptToken(encryptedToken string) (string, error) {
	return em.Decrypt(encryptedToken)
}

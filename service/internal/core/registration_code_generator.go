package core

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/speps/go-hashids/v2"
)

// RegistrationCodeGenerator 注册码生成器
type RegistrationCodeGenerator struct {
	hd *hashids.HashID
}

// NewRegistrationCodeGenerator 创建注册码生成器
func NewRegistrationCodeGenerator() (*RegistrationCodeGenerator, error) {
	hd, err := hashids.New()
	if err != nil {
		return nil, err
	}
	
	return &RegistrationCodeGenerator{
		hd: hd,
	}, nil
}

// GenerateCode 生成64位以上注册码
// 格式: RVCS-[时间戳]-[随机字符]-[校验码]
func (g *RegistrationCodeGenerator) GenerateCode() (string, error) {
	// 生成时间戳部分 (14字符)
	timestamp := time.Now().Format("20060102150405")
	
	// 生成随机字符部分 (32字符)
	randomChars := g.generateRandomString(32)
	
	// 构建基础字符串
	base := fmt.Sprintf("RVCS-%s-%s", timestamp, randomChars)
	
	// 生成校验码 (8字符)
	checksum := g.calculateChecksum(base)
	
	// 组合最终注册码
	registrationCode := fmt.Sprintf("%s-%08X", base, checksum)
	
	return registrationCode, nil
}

// generateRandomString 生成指定长度的随机字符串
func (g *RegistrationCodeGenerator) generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	// 生成随机字节
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// fallback to time-based if crypto fails
		for i := range bytes {
			bytes[i] = charset[int(time.Now().UnixNano())%len(charset)]
		}
		return string(bytes)
	}
	
	// 转换为字符集
	for i := range bytes {
		bytes[i] = charset[bytes[i]%byte(len(charset))]
	}
	
	return string(bytes)
}

// calculateChecksum 计算校验码
func (g *RegistrationCodeGenerator) calculateChecksum(data string) uint32 {
	// 使用简单的多项式哈希算法
	var hash uint32 = 5381
	for _, char := range data {
		hash = ((hash << 5) + hash) + uint32(char)
	}
	return hash
}

// ValidateCode 验证注册码格式
func (g *RegistrationCodeGenerator) ValidateCode(code string) bool {
	// 检查基本格式
	if len(code) < 58 { // 最小长度检查
		return false
	}
	
	// 检查前缀
	if len(code) < 4 || code[:4] != "RVCS" {
		return false
	}
	
	// 检查格式: RVCS-YYYYMMDDHHMMSS-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXX
	parts := []rune(code)
	if len(parts) < 58 {
		return false
	}
	
	// 简单的字符集检查
	validChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-"
	for _, char := range code {
		if char != '-' {
			valid := false
			for _, validChar := range validChars {
				if char == validChar {
					valid = true
					break
				}
			}
			if !valid {
				return false
			}
		}
	}
	
	return true
}

// ParseCode 解析注册码信息
func (g *RegistrationCodeGenerator) ParseCode(code string) (*ParsedRegistrationCode, error) {
	if !g.ValidateCode(code) {
		return nil, fmt.Errorf("invalid registration code format")
	}
	
	// 解析各个部分
	parts := splitByDelimiter(code, "-")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid registration code structure")
	}
	
	// 验证校验码
	base := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
	expectedChecksum := g.calculateChecksum(base)
	
	var actualChecksum uint32
	fmt.Sscanf(parts[3], "%X", &actualChecksum)
	
	if expectedChecksum != actualChecksum {
		return nil, fmt.Errorf("checksum verification failed")
	}
	
	return &ParsedRegistrationCode{
		Prefix:      parts[0],
		Timestamp:   parts[1],
		RandomPart:  parts[2],
		Checksum:    parts[3],
		FullCode:    code,
		CreatedTime: g.parseTimestamp(parts[1]),
	}, nil
}

// splitByDelimiter 按分隔符分割字符串
func splitByDelimiter(s, delimiter string) []string {
	var parts []string
	start := 0
	
	for i, char := range s {
		if string(char) == delimiter {
			if start < i {
				parts = append(parts, s[start:i])
			}
			start = i + 1
		}
	}
	
	// 添加最后一部分
	if start < len(s) {
		parts = append(parts, s[start:])
	}
	
	return parts
}

// parseTimestamp 解析时间戳
func (g *RegistrationCodeGenerator) parseTimestamp(timestampStr string) time.Time {
	// 时间格式: YYYYMMDDHHMMSS
	if len(timestampStr) != 14 {
		return time.Time{}
	}
	
	layout := "20060102150405"
	t, err := time.Parse(layout, timestampStr)
	if err != nil {
		return time.Time{}
	}
	
	return t
}

// ParsedRegistrationCode 解析后的注册码信息
type ParsedRegistrationCode struct {
	Prefix      string    `json:"prefix"`
	Timestamp   string    `json:"timestamp"`
	RandomPart  string    `json:"random_part"`
	Checksum    string    `json:"checksum"`
	FullCode    string    `json:"full_code"`
	CreatedTime time.Time `json:"created_time"`
}
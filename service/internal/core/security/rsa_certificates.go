package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"
)

// RSACertificates RSA证书管理器
type RSACertificates struct {
	ServerPrivateKey *rsa.PrivateKey
	ServerPublicKey  *rsa.PublicKey
}

// DeviceCertificate 设备证书
type DeviceCertificate struct {
	DeviceID     string    `json:"device_id"`
	PublicKeyPEM string    `json:"public_key_pem"`
	IssuedAt     time.Time `json:"issued_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	SerialNumber string    `json:"serial_number"`
	Signature    string    `json:"signature"`
}

// NewRSACertificates 创建RSA证书管理器
func NewRSACertificates(serverPrivateKeyPEM, serverPublicKeyPEM []byte) (*RSACertificates, error) {
	// 解析私钥和公钥...
	privateBlock, _ := pem.Decode(serverPrivateKeyPEM)
	publicBlock, _ := pem.Decode(serverPublicKeyPEM)
	
	privateKey, _ := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	publicKey, _ := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	
	return &RSACertificates{
		ServerPrivateKey: privateKey,
		ServerPublicKey:  publicKey.(*rsa.PublicKey),
	}, nil
}

// GenerateDeviceCertificate 生成设备证书
func (rc *RSACertificates) GenerateDeviceCertificate(
	deviceID string, 
	devicePubKey *rsa.PublicKey,
	validDays int,
) (*DeviceCertificate, error) {
	
	pubKeyPEM, _ := rc.GetPublicKeyPEM(devicePubKey)
	now := time.Now()
	
	cert := &DeviceCertificate{
		DeviceID:     deviceID,
		PublicKeyPEM: string(pubKeyPEM),
		IssuedAt:     now,
		ExpiresAt:    now.Add(time.Duration(validDays) * 24 * time.Hour),
		SerialNumber: generateSerialNumber(),
	}
	
	// 签名证书内容
	certData := fmt.Sprintf("%s|%s|%d|%d|%s", 
		cert.DeviceID, cert.PublicKeyPEM, 
		cert.IssuedAt.Unix(), cert.ExpiresAt.Unix(), cert.SerialNumber)
	
	signature, _ := rc.SignData([]byte(certData))
	cert.Signature = base64.StdEncoding.EncodeToString(signature)
	
	return cert, nil
}

// 辅助方法...
func (rc *RSACertificates) SignData(data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, rc.ServerPrivateKey, crypto.SHA256, hashed[:])
}

func (rc *RSACertificates) GetPublicKeyPEM(pubKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, _ := x509.MarshalPKIXPublicKey(pubKey)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubKeyBytes}), nil
}

func generateSerialNumber() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)
}
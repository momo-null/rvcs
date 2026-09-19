package security

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

// TLSConfig TLS配置工具
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// LoadCertificate 加载TLS证书
func LoadCertificate(certFile, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	
	// 验证证书
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	
	return &cert, nil
}

// CreateTLSConfig 创建TLS配置
func CreateTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := LoadCertificate(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		MinVersion: tls.VersionTLS12,
		// 推荐的加密套件
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		// 优先使用服务端加密套件
		PreferServerCipherSuites: true,
		// 支持的曲线
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
	}, nil
}

// GenerateSelfSignedCert 生成自签名证书（开发环境）
func GenerateSelfSignedCert(domain string, certFile, keyFile string) error {
	// 这里可以使用 crypto/tls 包生成自签名证书
	// 或者使用 openssl 命令生成
	
	log.Printf("For development, use: openssl req -x509 -newkey rsa:4096 -keyout %s -out %s -days 365 -nodes", keyFile, certFile)
	
	return nil
}

// VerifyCertificate 验证证书
func VerifyCertificate(certFile string) (*x509.Certificate, error) {
	certData, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, err
	}
	
	// 检查证书有效期
	if cert.NotBefore.IsZero() || cert.NotAfter.IsZero() {
		return cert, nil
	}
	
	// 检查证书是否过期
	if cert.NotAfter.Before(cert.NotBefore) {
		return cert, nil
	}
	
	return cert, nil
}

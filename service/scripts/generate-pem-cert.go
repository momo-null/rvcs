package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate-pem-cert.go <domain>")
		fmt.Println("Example: go run generate-pem-cert.go localhost")
		return
	}

	domain := os.Args[1]
	
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate private key: %v\n", err)
		return
	}

	// 创建证书模板
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"RVCS Development"},
			Country:       []string{"CN"},
			Province:      []string{"Development"},
			Locality:      []string{"Local"},
			CommonName:    domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain, "localhost"},
	}

	// 生成自签名证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Printf("Failed to create certificate: %v\n", err)
		return
	}

	// 保存证书文件
	certOut, err := os.Create("../certs/server.crt")
	if err != nil {
		fmt.Printf("Failed to open cert.pem for writing: %v\n", err)
		return
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		fmt.Printf("Failed to write data to cert.pem: %v\n", err)
		return
	}
	if err := certOut.Close(); err != nil {
		fmt.Printf("Error closing cert.pem: %v\n", err)
		return
	}
	fmt.Println("✅ Certificate saved to certs/server.crt")

	// 保存私钥文件
	keyOut, err := os.OpenFile("../certs/server.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Failed to open key.pem for writing: %v\n", err)
		return
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fmt.Printf("Unable to marshal private key: %v\n", err)
		return
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		fmt.Printf("Failed to write data to key.pem: %v\n", err)
		return
	}
	if err := keyOut.Close(); err != nil {
		fmt.Printf("Error closing key.pem: %v\n", err)
		return
	}
	fmt.Println("✅ Private key saved to certs/server.key")

	// 验证证书
	_, err = tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.key")
	if err != nil {
		fmt.Printf("❌ Certificate validation failed: %v\n", err)
		return
	}
	fmt.Println("✅ Certificate validated successfully!")
	fmt.Printf("📋 Certificate details:\n")
	fmt.Printf("   Domain: %s\n", domain)
	fmt.Printf("   Valid from: %s\n", template.NotBefore.Format("2006-01-02"))
	fmt.Printf("   Valid until: %s\n", template.NotAfter.Format("2006-01-02"))
	fmt.Printf("   Serial Number: %d\n", template.SerialNumber)
}
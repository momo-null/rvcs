package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: generate-cert <domain> [output-dir]")
		fmt.Println("Example: generate-cert localhost ./certs")
		os.Exit(1)
	}

	domain := os.Args[1]
	outputDir := "./certs"
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	// 确保输出目录存在
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// 创建证书模板
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"RVCS"},
			CommonName:   domain,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain, "localhost", "*.rvcs.local"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	// 生成证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	// 保存证书
	certOut, err := os.Create(fmt.Sprintf("%s/server.crt", outputDir))
	if err != nil {
		log.Fatalf("Failed to create certificate file: %v", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		log.Fatalf("Failed to write certificate: %v", err)
	}

	// 保存私钥
	keyOut, err := os.OpenFile(fmt.Sprintf("%s/server.key", outputDir), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to create private key file: %v", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}

	fmt.Printf("✓ Certificate and private key generated successfully!\n")
	fmt.Printf("  Certificate: %s/server.crt\n", outputDir)
	fmt.Printf("  Private Key: %s/server.key\n", outputDir)
	fmt.Printf("  Domain: %s\n", domain)
	fmt.Printf("  Valid until: %s\n", template.NotAfter.Format("2006-01-02"))
}

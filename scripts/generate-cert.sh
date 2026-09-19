#!/bin/bash

# HTTPS 证书生成脚本
# 用于开发环境生成自签名证书

set -e

DOMAIN=${1:-localhost}
OUTPUT_DIR=${2:-./certs}

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}======================================${NC}"
echo -e "${YELLOW}HTTPS Certificate Generator${NC}"
echo -e "${YELLOW}======================================${NC}"
echo ""

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 检查 openssl 是否安装
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}Error: openssl is not installed${NC}"
    exit 1
fi

# 生成私钥
echo -e "${GREEN}Generating private key...${NC}"
openssl genrsa -out "$OUTPUT_DIR/server.key" 2048

# 生成证书签名请求
echo -e "${GREEN}Generating certificate signing request...${NC}"
openssl req -new -key "$OUTPUT_DIR/server.key" \
    -out "$OUTPUT_DIR/server.csr" \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=RVCS/OU=Dev/CN=$DOMAIN"

# 生成自签名证书
echo -e "${GREEN}Generating self-signed certificate...${NC}"
openssl x509 -req -days 365 \
    -in "$OUTPUT_DIR/server.csr" \
    -signkey "$OUTPUT_DIR/server.key" \
    -out "$OUTPUT_DIR/server.crt"

# 删除 CSR 文件
rm "$OUTPUT_DIR/server.csr"

# 设置权限
chmod 600 "$OUTPUT_DIR/server.key"
chmod 644 "$OUTPUT_DIR/server.crt"

echo ""
echo -e "${GREEN}✓ Certificate and private key generated successfully!${NC}"
echo -e "${GREEN}  Certificate: ${OUTPUT_DIR}/server.crt${NC}"
echo -e "${GREEN}  Private Key: ${OUTPUT_DIR}/server.key${NC}"
echo -e "${GREEN}  Domain: $DOMAIN${NC}"
echo ""
echo -e "${YELLOW}Note: This is a self-signed certificate for development only.${NC}"
echo -e "${YELLOW}For production, use Let's Encrypt or purchase a certificate.${NC}"
echo ""

# 显示证书信息
echo -e "${YELLOW}Certificate Information:${NC}"
openssl x509 -in "$OUTPUT_DIR/server.crt" -text -noout | grep -E "(Subject:|Not Before|Not After|DNS:)" || true

echo ""
echo -e "${YELLOW}To use the certificate:${NC}"
echo -e "  1. Copy $OUTPUT_DIR/server.crt and $OUTPUT_DIR/server.key to your server"
echo -e "  2. Update the config file with the correct paths"
echo -e "  3. Restart the server"
echo ""
echo -e "${YELLOW}To trust the certificate in your browser:${NC}"
echo -e "  1. Open https://$DOMAIN in your browser"
echo -e "  2. Click 'Advanced' -> 'Proceed to $DOMAIN (unsafe)'"
echo ""

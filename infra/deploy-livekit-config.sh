#!/bin/bash

# ========================================
# 配置服务器IP的脚本
# Remote Vision & Control System
# ========================================

set -e

# 默认IP
DEFAULT_IP="YOUR_SERVER_IP"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "RVCS 服务器IP配置"
echo "=========================================="
echo ""

# 检查参数
if [ -n "$1" ]; then
    SERVER_IP="$1"
else
    echo -e "${YELLOW}请输入服务器IP地址 (回车使用默认: $DEFAULT_IP):${NC}"
    read -r INPUT_IP
    if [ -z "$INPUT_IP" ]; then
        SERVER_IP="$DEFAULT_IP"
    else
        SERVER_IP="$INPUT_IP"
    fi
fi

echo ""
echo "配置服务器IP: $SERVER_IP"
echo ""

# 修改 service/configs/config.yaml 中的 livekit.server_url
if [ -f "service/configs/config.yaml" ]; then
    echo "正在更新 service/configs/config.yaml..."
    sed -i "s|server_url:.*|server_url: \"ws://$SERVER_IP:7888\"|g" service/configs/config.yaml
    echo -e "${GREEN}✓ service/configs/config.yaml 已更新${NC}"
else
    echo -e "${YELLOW}⚠ service/configs/config.yaml 不存在${NC}"
fi

# 修改 livekit.yaml（如果有）
if [ -f "livekit.yaml" ]; then
    echo "正在更新 livekit.yaml..."
    # livekit.yaml 不需要修改IP，因为它在容器内部
    echo -e "${GREEN}✓ livekit.yaml 无需修改${NC}"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}配置完成！服务器IP: $SERVER_IP${NC}"
echo "=========================================="
echo ""
echo "现在可以运行部署脚本："
echo "  ./deploy.sh"
echo ""

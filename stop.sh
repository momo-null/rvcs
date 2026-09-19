#!/bin/bash

# ========================================
# RVCS 服务停止脚本
# Remote Vision & Control System
# ========================================

set -e

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================="
echo "RVCS 服务停止"
echo "=========================================="
echo ""
# 停止后端服务
echo "正在停止后端服务..."

if pgrep -f "rvcs-server" > /dev/null; then
    echo "停止后端服务..."
    kill $(cat rvcs-server.pid 2>/dev/null || pgrep -f 'rvcs-server') 2>/dev/null || true
    
    # 等待进程结束
    sleep 2
    
    # 检查是否还有残留进程
    if pgrep -f "rvcs-server" > /dev/null; then
        echo -e "${YELLOW}强制停止残留进程...${NC}"
        pkill -9 -f "rvcs-server" || true
    fi
    
    echo -e "${GREEN}✓ 后端服务已停止${NC}"
else
    echo -e "${YELLOW}后端服务未运行${NC}"
fi

echo ""
echo "=========================================="
echo -e "${GREEN}服务已停止！${NC}"
echo "=========================================="
echo ""
echo "管理命令："
echo "  - 启动所有服务: ./start.sh"
echo "  - 查看后端日志: tail -f logs/rvcs-server.log"

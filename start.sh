#!/bin/bash

# ========================================
# RVCS 服务启动脚本
# Remote Vision & Control System
# ========================================

set -e

# 创建 logs 目录（如果不存在）
if [ ! -d "logs" ]; then
    echo "创建 logs 目录..."
    mkdir -p logs
fi

# 设置可执行权限（如果用户没有权限）
chmod 777 ./rvcs-server 2>/dev/null || {
    echo -e "\033[0;33m警告: 无法设置 rvcs-server 可执行权限\033[0m"
    echo -e "\033[0;33m提示: 请使用 sudo 运行此脚本，或手动执行: chmod 777 rvcs-server\033[0m"
    # 尝试使用 bash 直接执行
    exec bash -c "./rvcs-server > logs/rvcs-server.log 2>&1 &"
}

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "=========================================="
echo "RVCS 服务启动"
echo "=========================================="
echo ""

# 启动后端服务
echo "正在启动后端服务..."

# 检查后端服务是否已运行
if pgrep -f "rvcs-server" > /dev/null; then
    echo -e "${YELLOW}后端服务已在运行，PID: $(pgrep -f 'rvcs-server')${NC}"
else
    echo "启动后端服务..."
    nohup ./rvcs-server > logs/rvcs-server.log 2>&1 &
    echo $! > rvcs-server.pid

    # 等待服务启动
    echo "等待服务启动..."
    sleep 3

    # 检查服务是否启动成功
    if pgrep -f "rvcs-server" > /dev/null; then
        echo -e "${GREEN}✓ 后端服务启动成功，PID: $(cat rvcs-server.pid)${NC}"
    else
        echo -e "${RED}✗ 后端服务启动失败${NC}"
        echo "查看日志: tail -f logs/rvcs-server.log"
        exit 1
    fi
fi

echo ""
echo "=========================================="
echo -e "${GREEN}启动完成！${NC}"
echo "=========================================="
echo ""
echo "管理命令："
echo "  - 查看后端日志: tail -f logs/rvcs-server.log"
echo "  - 查看 docker-compose 日志: docker-compose -f docker-compose-livekit.yml logs -f"
echo "  - 停止所有服务: ./stop.sh"

#!/bin/bash

# ========================================
# RVCS 部署检查脚本
# Remote Vision & Control System
# ========================================

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "RVCS 部署状态检查"
echo "=========================================="
echo ""

# 检查计数器
PASS=0
FAIL=0
WARN=0

# 检查函数
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASS++))
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAIL++))
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
    ((WARN++))
}

# 1. 检查 Docker
echo "=========================================="
echo "1. Docker 检查"
echo "=========================================="

if command -v docker &> /dev/null; then
    check_pass "Docker 已安装"
    DOCKER_VERSION=$(docker --version)
    echo "  版本: $DOCKER_VERSION"
else
    check_fail "Docker 未安装"
fi

if command -v docker-compose &> /dev/null; then
    check_pass "Docker Compose 已安装"
    COMPOSE_VERSION=$(docker-compose --version)
    echo "  版本: $COMPOSE_VERSION"
else
    check_fail "Docker Compose 未安装"
fi
echo ""

# 2. 检查 LiveKit 容器
echo "=========================================="
echo "2. LiveKit 容器检查"
echo "=========================================="

if docker ps --format '{{.Names}}' | grep -q "^rvcs-livekit$"; then
    check_pass "LiveKit 容器正在运行"

    # 检查端口
    if netstat -tuln 2>/dev/null | grep -q ":7888 "; then
        check_pass "LiveKit 端口 7888 已开放"
    else
        check_fail "LiveKit 端口 7888 未开放"
    fi

    # 检查容器日志（最近 10 行）
    echo ""
    echo "  最近 10 行日志:"
    docker logs --tail 10 rvcs-livekit 2>&1 | sed 's/^/    /'
else
    check_fail "LiveKit 容器未运行"
fi
echo ""

# 3. 检查后端服务
echo "=========================================="
echo "3. 后端服务检查"
echo "=========================================="

if [ -f "/opt/rvcs/rvcs-server" ]; then
    check_pass "后端可执行文件存在"
else
    check_fail "后端可执行文件不存在"
fi

if pgrep -f "rvcs-server" > /dev/null; then
    check_pass "后端服务正在运行"
    SERVER_PID=$(pgrep -f "rvcs-server")
    echo "  PID: $SERVER_PID"

    # 检查端口
    if netstat -tuln 2>/dev/null | grep -q ":28443 "; then
        check_pass "后端 HTTPS 端口 28443 已开放"
    else
        check_fail "后端 HTTPS 端口 28443 未开放"
    fi
else
    check_fail "后端服务未运行"
fi

if [ -f "/opt/rvcs/rvcs-server.pid" ]; then
    check_pass "PID 文件存在"
else
    check_warn "PID 文件不存在"
fi
echo ""

# 4. 检查配置文件
echo "=========================================="
echo "4. 配置文件检查"
echo "=========================================="

if [ -f "/opt/rvcs/configs/config.yaml" ]; then
    check_pass "后端配置文件存在"

    # 提取关键配置
    HTTPS_PORT=$(grep "https_port:" /opt/rvcs/configs/config.yaml | awk '{print $2}')
    LIVEKIT_URL=$(grep "server_url:" /opt/rvcs/configs/config.yaml | awk '{print $2}' | tr -d '"')
    echo "  HTTPS 端口: $HTTPS_PORT"
    echo "  LiveKit 地址: $LIVEKIT_URL"
else
    check_fail "后端配置文件不存在"
fi
echo ""

# 5. 检查日志文件
echo "=========================================="
echo "5. 日志文件检查"
echo "=========================================="

if [ -f "/opt/rvcs/logs/rvcs-server.log" ]; then
    check_pass "后端日志文件存在"

    # 检查日志文件大小
    LOG_SIZE=$(du -h /opt/rvcs/logs/rvcs-server.log | awk '{print $1}')
    echo "  日志大小: $LOG_SIZE"

    # 检查最近日志
    echo ""
    echo "  最近 10 行日志:"
    tail -10 /opt/rvcs/logs/rvcs-server.log 2>/dev/null | sed 's/^/    /'
else
    check_warn "后端日志文件不存在"
fi
echo ""

# 6. 检查端口占用
echo "=========================================="
echo "6. 端口占用检查"
echo "=========================================="

PORTS=("28443" "7888")
for PORT in "${PORTS[@]}"; do
    if netstat -tuln 2>/dev/null | grep -q ":$PORT "; then
        check_pass "端口 $PORT 已监听"
        PROCESS=$(netstat -tuln 2>/dev/null | grep ":$PORT " | awk '{print $7}' | cut -d'/' -f2)
        echo "  进程: $PROCESS"
    else
        check_fail "端口 $PORT 未监听"
    fi
done
echo ""

# 7. 检查防火墙
echo "=========================================="
echo "7. 防火墙检查"
echo "=========================================="

if command -v ufw &> /dev/null; then
    echo "  防火墙类型: UFW"
    UFW_STATUS=$(ufw status | head -1)
    echo "  状态: $UFW_STATUS"

    if ufw status | grep -q "28443/tcp"; then
        check_pass "端口 28443 已开放"
    else
        check_warn "端口 28443 可能未开放"
    fi

    if ufw status | grep -q "7888/tcp"; then
        check_pass "端口 7888 已开放"
    else
        check_warn "端口 7888 可能未开放"
    fi
elif command -v firewall-cmd &> /dev/null; then
    echo "  防火墙类型: firewalld"
    FIREWALL_STATUS=$(firewall-cmd --state 2>/dev/null)
    echo "  状态: $FIREWALL_STATUS"

    if firewall-cmd --list-ports | grep -q "28443/tcp"; then
        check_pass "端口 28443 已开放"
    else
        check_warn "端口 28443 可能未开放"
    fi

    if firewall-cmd --list-ports | grep -q "7888/tcp"; then
        check_pass "端口 7888 已开放"
    else
        check_warn "端口 7888 可能未开放"
    fi
else
    check_warn "未检测到防火墙工具 (ufw 或 firewalld)"
fi
echo ""

# 8. 系统资源检查
echo "=========================================="
echo "8. 系统资源检查"
echo "=========================================="

# CPU 使用率
CPU_USAGE=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
echo "  CPU 使用率: ${CPU_USAGE}%"

# 内存使用率
MEM_USAGE=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
echo "  内存使用率: ${MEM_USAGE}%"

# 磁盘使用率
DISK_USAGE=$(df -h / | tail -1 | awk '{print $5}')
echo "  磁盘使用率: $DISK_USAGE"

if (( $(echo "$MEM_USAGE < 90" | bc -l) )); then
    check_pass "内存使用正常"
else
    check_warn "内存使用率过高"
fi
echo ""

# 9. 网络连接检查
echo "=========================================="
echo "9. 网络连接检查"
echo "=========================================="

# 获取服务器公网 IP
PUBLIC_IP=$(hostname -I | awk '{print $1}')
echo "  服务器 IP: $PUBLIC_IP"

# 测试 HTTPS 连接
if command -v curl &> /dev/null; then
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -k https://localhost:28443 2>/dev/null || echo "000")
    if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "301" ] || [ "$HTTP_CODE" = "302" ]; then
        check_pass "HTTPS 服务可访问 (HTTP $HTTP_CODE)"
    else
        check_warn "HTTPS 服务可能不可访问 (HTTP $HTTP_CODE)"
    fi
else
    check_warn "curl 未安装，无法测试 HTTPS 连接"
fi
echo ""

# 总结
echo "=========================================="
echo "检查总结"
echo "=========================================="
echo ""
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo -e "警告: ${YELLOW}$WARN${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}所有关键检查通过！${NC}"
    echo ""
    echo "访问地址: https://$PUBLIC_IP:28443"
    echo "默认账户: admin / admin123"
    exit 0
else
    echo -e "${RED}发现 $FAIL 个问题，请检查上述失败项${NC}"
    exit 1
fi

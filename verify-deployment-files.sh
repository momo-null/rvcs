#!/bin/bash

# ========================================
# 验证部署文件完整性
# ========================================

echo "=========================================="
echo "验证部署文件完整性"
echo "=========================================="
echo ""

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

PASS=0
FAIL=0

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
}

echo "检查部署脚本..."
echo ""

# 部署脚本
if [ -f "deploy.sh" ]; then
    check_pass "deploy.sh 存在"
    chmod +x deploy.sh
else
    check_fail "deploy.sh 不存在"
fi

if [ -f "infra/deploy-livekit-config.sh" ]; then
    check_pass "infra/deploy-livekit-config.sh 存在"
    chmod +x infra/deploy-livekit-config.sh
elif [ -f "deploy-livekit-config.sh" ]; then
    check_pass "deploy-livekit-config.sh 存在"
    chmod +x deploy-livekit-config.sh
else
    check_fail "deploy-livekit-config.sh 不存在"
fi

if [ -f "infra/check-deployment.sh" ]; then
    check_pass "infra/check-deployment.sh 存在"
    chmod +x infra/check-deployment.sh
elif [ -f "check-deployment.sh" ]; then
    check_pass "check-deployment.sh 存在"
    chmod +x check-deployment.sh
else
    check_fail "check-deployment.sh 不存在"
fi

echo ""
echo "检查配置文件..."
echo ""

# 配置文件
if [ -f "infra/docker-compose-livekit.yml" ]; then
    check_pass "infra/docker-compose-livekit.yml 存在"
elif [ -f "docker-compose-livekit.yml" ]; then
    check_pass "docker-compose-livekit.yml 存在"
else
    check_fail "docker-compose-livekit.yml 不存在"
fi

if [ -f "infra/livekit.yaml" ]; then
    check_pass "infra/livekit.yaml 存在"
elif [ -f "livekit.yaml" ]; then
    check_pass "livekit.yaml 存在"
else
    check_fail "livekit.yaml 不存在"
fi

if [ -f "service/configs/config.yaml" ]; then
    check_pass "service/configs/config.yaml 存在"
else
    check_fail "service/configs/config.yaml 不存在"
fi

echo ""
echo "检查文档文件..."
echo ""

# 文档文件
if [ -f "QUICK-DEPLOYMENT.md" ]; then
    check_pass "QUICK-DEPLOYMENT.md 存在"
else
    check_fail "QUICK-DEPLOYMENT.md 不存在"
fi

if [ -f "README-DEPLOYMENT.md" ]; then
    check_pass "README-DEPLOYMENT.md 存在"
else
    check_fail "README-DEPLOYMENT.md 不存在"
fi

if [ -f "DEPLOYMENT-INSTRUCTIONS.md" ]; then
    check_pass "DEPLOYMENT-INSTRUCTIONS.md 存在"
else
    check_fail "DEPLOYMENT-INSTRUCTIONS.md 不存在"
fi

if [ -f "DEPLOYMENT-PACKAGE-SUMMARY.md" ]; then
    check_pass "DEPLOYMENT-PACKAGE-SUMMARY.md 存在"
else
    check_warn "DEPLOYMENT-PACKAGE-SUMMARY.md 不存在（可选）"
fi

echo ""
echo "检查构建脚本..."
echo ""

if [ -f "build-all.sh" ]; then
    check_pass "build-all.sh 存在"
    chmod +x build-all.sh
else
    check_warn "build-all.sh 不存在（Windows 可使用 .bat）"
fi

if [ -f "build-all.bat" ]; then
    check_pass "build-all.bat 存在"
else
    check_warn "build-all.bat 不存在（Linux 可使用 .sh）"
fi

echo ""
echo "=========================================="
echo "验证结果"
echo "=========================================="
echo ""
echo -e "通过: ${GREEN}$PASS${NC}"
echo -e "失败: ${RED}$FAIL${NC}"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}所有必需文件检查通过！${NC}"
    echo ""
    echo "现在可以执行构建："
    echo "  Windows: build-all.bat"
    echo "  Linux/Mac: ./build-all.sh"
    echo ""
    exit 0
else
    echo -e "${RED}发现 $FAIL 个缺失文件，请检查上述失败项${NC}"
    exit 1
fi

#!/bin/bash

# 初始化管理员用户脚本

echo "======================================"
echo "初始化管理员用户"
echo "======================================"

cd "$(dirname "$0")/.."

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境"
    exit 1
fi

# 检查配置文件是否存在
if [ ! -f "configs/config.yaml" ]; then
    echo "错误: 配置文件不存在 (configs/config.yaml)"
    exit 1
fi

# 运行初始化脚本
go run scripts/init_admin_user.go

echo ""
echo "======================================"
echo "初始化完成"
echo "======================================"

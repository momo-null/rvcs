@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: ========================================
:: RVCS Web Build Script
:: 前端构建并集成到后端服务
:: ========================================

echo ========================================
echo RVCS Web Build Script
echo ========================================
echo.

:: 清理历史构建产物
echo [1/5] 清理历史构建产物...
if exist "dist" (
    echo 删除 dist 目录...
    rmdir /s /q "dist" 2>nul
    if exist "dist" (
        echo 警告: 无法删除 dist 目录，请检查文件是否被占用
        timeout /t 2 >nul
    ) else (
        echo dist 目录已删除
    )
) else (
    echo dist 目录不存在，跳过清理
)

:: 清理 service/web 目录中的旧文件
echo.
echo [2/5] 清理 service/web 旧文件...
if exist "..\service\web" (
    echo 删除 service\web 目录...
    rmdir /s /q "..\service\web" 2>nul
    if exist "..\service\web" (
        echo 警告: 无法删除 service\web 目录，请检查文件是否被占用
        timeout /t 2 >nul
    ) else (
        echo service\web 目录已删除
    )
) else (
    echo service\web 目录不存在，跳过清理
)

:: 检查 node_modules 是否存在
echo.
echo [3/5] 检查依赖...
if not exist "node_modules" (
    echo 安装依赖...
    call npm install
    if errorlevel 1 (
        echo 错误: 依赖安装失败！
        pause
        exit /b 1
    )
) else (
    echo 依赖已存在，跳过安装
)

:: 构建前端
echo.
echo [4/5] 构建前端项目...
call npm run build
if errorlevel 1 (
    echo 错误: 前端构建失败！
    pause
    exit /b 1
)

:: 复制到 service/web
echo.
echo [5/5] 复制构建产物到 service/web...

:: 确保目标目录存在
if not exist "..\service\web" (
    mkdir "..\service\web"
)

:: 使用 robocopy 复制文件
robocopy "dist" "..\service\web" /E /NFL /NDL /NJH /NJS /NP

if %errorlevel% leq 7 (
    echo 构建成功！文件已复制到 ..\service\web\
) else (
    echo 错误: 文件复制失败！
    pause
    exit /b 1
)

echo.
echo ========================================
echo 构建完成！
echo ========================================
echo.
echo 后续步骤:
echo   1. cd ..\service
echo   2. go run cmd/server/main.go
echo   3. 访问: https://localhost:28443
echo.

timeout /t 3 >nul

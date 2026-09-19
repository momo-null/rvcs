@echo off
setlocal enabledelayedexpansion

echo ========================================
echo Android应用重新编译安装脚本
echo ========================================

:: 设置环境变量
set ANDROID_HOME=D:\Android\Sdk
set JAVA_HOME=C:\Program Files\Java\jdk-17
set PATH=%ANDROID_HOME%\platform-tools;%ANDROID_HOME%\tools;%JAVA_HOME%\bin;%PATH%

echo 环境变量设置:
echo ANDROID_HOME=%ANDROID_HOME%
echo JAVA_HOME=%JAVA_HOME%
echo.

:: 检查必要的工具
echo 检查必要工具...
where java >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到Java
    goto :error
)

where adb >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到ADB
    goto :error
)

echo 工具检查通过
echo.

:: 检查设备连接
echo 检查Android设备连接...
adb devices
echo.

:: 进入项目目录
cd /d D:\code\rvcs\android

:: 清理之前的构建
echo 清理之前的构建...
if exist app\build rmdir /s /q app\build
echo 清理完成
echo.

:: 初始化Gradle wrapper（如果不存在）
if not exist gradlew.bat (
    echo 创建Gradle wrapper...
    gradle wrapper --gradle-version 7.4
    if %ERRORLEVEL% NEQ 0 (
        echo 警告: Gradle wrapper创建失败，尝试直接使用系统Gradle
    )
)

:: 构建APK
echo 开始构建APK...
if exist gradlew.bat (
    call gradlew.bat assembleDebug
) else (
    gradle assembleDebug
)

if %ERRORLEVEL% NEQ 0 (
    echo 构建失败，请检查错误信息
    goto :error
)

:: 检查APK是否生成
if not exist "app\build\outputs\apk\debug\app-debug.apk" (
    echo 错误: APK文件未生成
    goto :error
)

echo APK构建成功！

:: 安装到设备
echo 正在安装到Android设备...
adb install -r "app\build\outputs\apk\debug\app-debug.apk"

if %ERRORLEVEL% EQU 0 (
    echo ========================================
    echo 应用安装成功！
    echo ========================================
    echo 应用包名: com.rvcs.android
    echo 主要功能:
    echo - 设备注册和认证
    echo - 摄像头服务管理  
    echo - 实时视频流传输
    echo ========================================
) else (
    echo 安装失败，请检查设备连接和权限设置
    goto :error
)

goto :success

:error
echo ========================================
echo 操作失败！
echo ========================================
echo 请检查:
echo 1. Android SDK是否正确安装
echo 2. 设备是否已连接并开启USB调试
echo 3. 是否有足够的存储空间
echo 4. 构建日志中的具体错误信息
echo ========================================
exit /b 1

:success
echo 操作完成！
pause
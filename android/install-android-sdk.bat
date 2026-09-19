@echo off
echo ========================================
echo Android SDK 自动安装脚本
echo ========================================

REM 设置安装目录
set INSTALL_DIR=D:\Android\Sdk
set CMD_TOOLS_URL=https://dl.google.com/android/repository/commandlinetools-win-9477386_latest.zip

echo.
echo 1. 创建安装目录: %INSTALL_DIR%
mkdir "%INSTALL_DIR%" 2>nul

echo.
echo 2. 下载Android Command-line Tools...
powershell -Command "Invoke-WebRequest -Uri '%CMD_TOOLS_URL%' -OutFile '%TEMP%\cmd-tools.zip'"

echo.
echo 3. 解压工具包...
powershell -Command "Expand-Archive -Path '%TEMP%\cmd-tools.zip' -DestinationPath '%INSTALL_DIR%' -Force"

echo.
echo 4. 重命名tools目录...
ren "%INSTALL_DIR%\cmdline-tools" "tools" 2>nul

echo.
echo 5. 设置环境变量...
setx ANDROID_HOME "%INSTALL_DIR%"
setx PATH "%PATH%;%INSTALL_DIR%\tools\bin;%INSTALL_DIR%\platform-tools"

echo.
echo 6. 安装必需的SDK组件...
echo 这可能需要几分钟时间...

echo 使用阿里云镜像加速下载...
set HTTP_PROXY=mirrors.aliyun.com
"%INSTALL_DIR%\tools\bin\sdkmanager.bat" --sdk_root="%INSTALL_DIR%" --proxy=http --proxy_host=mirrors.aliyun.com --proxy_port=80 "platform-tools" "platforms;android-34" "build-tools;34.0.0"

echo.
echo ========================================
echo 安装完成！
echo SDK路径: %INSTALL_DIR%
echo 环境变量 ANDROID_HOME 已设置
echo 请重启命令提示符使环境变量生效
echo ========================================

pause
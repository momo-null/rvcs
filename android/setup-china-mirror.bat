@echo off
echo ========================================
echo Android SDK 国内镜像配置脚本
echo ========================================

REM 检测Android SDK路径
set SDK_PATH=
if exist "%ANDROID_HOME%" (
    set SDK_PATH=%ANDROID_HOME%
) else if exist "D:\Android\Sdk" (
    set SDK_PATH=D:\Android\Sdk
) else if exist "%LOCALAPPDATA%\Android\Sdk" (
    set SDK_PATH=%LOCALAPPDATA%\Android\Sdk
)

if "%SDK_PATH%"=="" (
    echo 错误：未找到Android SDK，请先安装SDK
    echo 运行 install-android-sdk.bat 进行安装
    pause
    exit /b 1
)

echo 检测到SDK路径: %SDK_PATH%

REM 创建镜像配置文件
echo.
echo 配置国内镜像源...

set REPO_CFG=%SDK_PATH%\.android\repositories.cfg
mkdir "%SDK_PATH%\.android" 2>nul

(
echo ### User Sources for Android SDK Manager
echo count=3
echo 0.id=aliyun
echo 0.url=https://mirrors.aliyun.com/android/repository/
echo 1.id=tsinghua  
echo 1.url=https://mirrors.tuna.tsinghua.edu.cn/help/AOSP/
echo 2.id=huawei
echo 2.url=https://repo.huaweicloud.com/repository/maven/
) > "%REPO_CFG%"

echo 镜像配置文件已创建: %REPO_CFG%

REM 配置Gradle镜像（如果存在gradle.properties）
set GRADLE_PROPS=%USERPROFILE%\.gradle\gradle.properties
if not exist "%USERPROFILE%\.gradle" mkdir "%USERPROFILE%\.gradle"

echo.
echo 配置Gradle全局镜像...

(
echo # 国内镜像配置
echo org.gradle.jvmargs=-Xmx2048m -XX:MaxMetaspaceSize=512m
echo org.gradle.parallel=true
echo org.gradle.caching=true
echo # 阿里云镜像
echo systemProp.http.proxyHost=mirrors.aliyun.com
echo systemProp.https.proxyHost=mirrors.aliyun.com
) >> "%GRADLE_PROPS%"

echo Gradle配置已更新

echo.
echo ========================================
echo 镜像配置完成！
echo 推荐使用以下命令安装SDK组件：
echo sdkmanager --proxy=http --proxy_host=mirrors.tuna.tsinghua.edu.cn --proxy_port=80 "platform-tools" "platforms;android-34" "build-tools;34.0.0"
echo ========================================

pause
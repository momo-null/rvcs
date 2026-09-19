@echo off
echo ========================================
echo Android SDK 完整配置脚本
echo ========================================

REM 设置环境变量
set ANDROID_HOME=D:\Android\Sdk
set PATH=%PATH%;%ANDROID_HOME%\tools\bin;%ANDROID_HOME%\platform-tools

echo SDK路径: %ANDROID_HOME%

REM 创建许可证目录
mkdir "%ANDROID_HOME%\licenses" 2>nul

REM 创建许可证文件
echo 8933bad161af4178b1185d1a37fbf41ea5269c55 > "%ANDROID_HOME%\licenses\android-sdk-license"
echo d975f751698a77b662f1254ddbeed3901e976f5a > "%ANDROID_HOME%\licenses\intel-android-extra-license"

echo 许可证文件已创建

REM 使用sdkmanager安装必需组件
echo 正在安装Android SDK组件...
"%ANDROID_HOME%\tools\bin\sdkmanager.bat" --sdk_root="%ANDROID_HOME%" "platform-tools" "platforms;android-34" "build-tools;30.0.3"

echo.
echo 配置完成！现在可以构建项目了：
echo cd D:\code\rvcs\android
echo gradlew.bat assembleDebug

pause
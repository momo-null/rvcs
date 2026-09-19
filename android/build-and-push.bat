@echo off
echo 正在构建Android应用...

:: 设置Android SDK路径
set ANDROID_HOME=D:\Android\Sdk
set PATH=%ANDROID_HOME%\platform-tools;%PATH%

:: 检查是否有连接的设备
echo 检查连接的设备...
adb devices

:: 使用Gradle构建APK
echo 开始构建APK...
gradle assembleDebug

:: 检查APK是否生成成功
if exist "app\build\outputs\apk\debug\app-debug.apk" (
    echo APK构建成功！
    echo 正在安装到设备...
    adb install -r app\build\outputs\apk\debug\app-debug.apk
    if %ERRORLEVEL% EQU 0 (
        echo 应用安装成功！
    ) else (
        echo 应用安装失败，请检查设备连接。
    )
) else (
    echo APK构建失败，请检查错误信息。
)

pause
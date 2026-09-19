@echo off
setlocal enabledelayedexpansion

:: ========================================
:: RVCS Android Build Script (Windows)
:: Remote Vision & Control System - Android App
:: ========================================

echo ========================================
echo RVCS Android Build Script
echo Remote Vision & Control System - Android App
echo ========================================
echo.

:: 设置变量
set ANDROID_DIR=android
set BUILD_DIR=dist
set TIMESTAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%

:: 检查Android目录是否存在
if not exist "%ANDROID_DIR%" (
    echo ERROR: Android directory not found!
    echo Please make sure the 'android' folder exists in the project root.
    exit /b 1
)

echo [1/4] Checking Android project structure...
cd "%ANDROID_DIR%"

:: 检查必要的文件
if not exist "gradlew.bat" (
    echo ERROR: gradlew.bat not found!
    echo Please make sure this is a valid Android project.
    cd ..
    exit /b 1
)

if not exist "app" (
    echo ERROR: app directory not found!
    echo Please make sure this is a valid Android project.
    cd ..
    exit /b 1
)

cd ..

:: 清理旧的Android构建输出
echo.
echo [2/4] Cleaning previous Android builds...
if exist "%BUILD_DIR%\app" (
    rd /s /q "%BUILD_DIR%\app"
)

:: 构建Android应用
echo.
echo [3/4] Building Android app (Release)...
cd "%ANDROID_DIR%"

echo Running Gradle build...
call gradlew.bat assembleRelease
if errorlevel 1 (
    echo ERROR: Android build failed!
    echo Trying debug build instead...
    call gradlew.bat assembleDebug
    if errorlevel 1 (
        echo ERROR: Both release and debug builds failed!
        cd ..
        exit /b 1
    ) else (
        set BUILD_TYPE=debug
        echo Successfully built debug version
    )
) else (
    set BUILD_TYPE=release
    echo Successfully built release version
)

cd ..

:: 复制APK到dist目录
echo.
echo [4/4] Copying APK to distribution directory...
if not exist "%BUILD_DIR%\app" (
    mkdir "%BUILD_DIR%\app"
)

if "%BUILD_TYPE%"=="release" (
    set APK_SOURCE=%ANDROID_DIR%\app\build\outputs\apk\release\app-release.apk
    set APK_DEST=%BUILD_DIR%\app\rvcs-app-release.apk
) else (
    set APK_SOURCE=%ANDROID_DIR%\app\build\outputs\apk\debug\app-debug.apk
    set APK_DEST=%BUILD_DIR%\app\rvcs-app-debug.apk
)

if exist "%APK_SOURCE%" (
    copy /y "%APK_SOURCE%" "%APK_DEST%"
    if errorlevel 1 (
        echo WARNING: Failed to copy APK file
    ) else (
        echo APK copied successfully: %APK_DEST%
    )
) else (
    echo WARNING: APK file not found at expected location
    echo Looking for APK files in build directory...
    dir "%ANDROID_DIR%\app\build\outputs\apk" /s /b | findstr "\.apk$"
)

:: 完成
echo.
echo ========================================
echo Android build completed!
echo ========================================
echo.
echo Output directory: %BUILD_DIR%\app\
echo Build type: %BUILD_TYPE%
echo.
echo To install the app on an Android device:
echo   1. Enable USB debugging on your device
echo   2. Connect device via USB
echo   3. Run: adb install %APK_DEST:\=/%
echo.

endlocal
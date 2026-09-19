@echo off
setlocal enabledelayedexpansion

:: ========================================
:: RVCS Build Script (Windows)
:: Remote Vision & Control System
:: ========================================

echo ========================================
echo RVCS Build Script
echo Remote Vision & Control System
echo ========================================
echo.

:: 璁剧疆鍙橀噺
set BUILD_DIR=dist
set SERVICE_DIR=service
set WEB_DIR=web
set ANDROID_DIR=android
set INFRA_DIR=infra
set TIMESTAMP=%date:~0,4%%date:~5,2%%date:~8,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%

:: 娓呯悊鏃х殑鏋勫缓鐩綍
echo [1/6] Cleaning build directory...
if exist "%BUILD_DIR%" (
    rd /s /q "%BUILD_DIR%"
)
if not exist "%BUILD_DIR%" (
    mkdir "%BUILD_DIR%"
)

:: 鏋勫缓Web绔?
echo.
echo [2/6] Building Web frontend (force rebuild)...
cd "%WEB_DIR%"

:: 娓呯悊Vite缂撳瓨鍜屾棫鐨勬瀯寤轰骇鐗?
echo Cleaning Vite cache and old build...
if exist "node_modules\.vite" (
    rd /s /q "node_modules\.vite"
)
if exist "dist" (
    rd /s /q "dist"
)

call npm install
:: 鍒犻櫎缂撳瓨鍚庣洿鎺ユ瀯寤猴紝Vite浼氳嚜鍔ㄩ噸鏂扮紪璇戞墍鏈夋枃浠?
call npm run build
if errorlevel 1 (
    echo ERROR: Web build failed!
    cd ..
    exit /b 1
)
cd ..

:: 灏哤eb鏋勫缓浜х墿澶嶅埗鍒癝ervice
echo.
echo [3/6] Integrating web files into service...
if not exist "%SERVICE_DIR%\web" (
    mkdir "%SERVICE_DIR%\web"
)
xcopy /s /e /y /q "%WEB_DIR%\dist\*" "%SERVICE_DIR%\web\" >nul 2>&1
if errorlevel 1 (
    echo ERROR: Failed to integrate web files!
    exit /b 1
)

:: 鏋勫缓Service绔?
echo.
echo [4/6] Building Service backend with integrated web frontend...
cd "%SERVICE_DIR%"
echo Downloading Go dependencies...
go mod tidy
if errorlevel 1 (
    echo ERROR: Go mod tidy failed!
    exit /b 1
)

echo Building RVCS server...
go build -tags purego -v -o ..\dist\rvcs-server.exe -ldflags="-s -w" ./cmd/server
if errorlevel 1 (
    echo ERROR: Build failed!
    exit /b 1
)

cd ..

:: 澶嶅埗Service鍒癲ist鐩綍
echo.
echo [5/6] Copying integrated service files to dist...
if not exist "%BUILD_DIR%\service" (
    mkdir "%BUILD_DIR%\service"
)

:: 澶嶅埗鍙墽琛屾枃浠?
copy /y "%BUILD_DIR%\rvcs-server.exe" "%BUILD_DIR%\service\"

:: 澶嶅埗web闈欐€佹枃浠跺埌dist/service
if exist "%SERVICE_DIR%\web" (
    if not exist "%BUILD_DIR%\service\web" (
        mkdir "%BUILD_DIR%\service\web"
    )
    xcopy /s /e /y /q "%SERVICE_DIR%\web\*" "%BUILD_DIR%\service\web\" >nul 2>&1
)

:: 澶嶅埗閰嶇疆鏂囦欢
if exist "%SERVICE_DIR%\configs" (
    if not exist "%BUILD_DIR%\service\configs" (
        mkdir "%BUILD_DIR%\service\configs"
    )
    xcopy /s /e /y /q "%SERVICE_DIR%\configs\*" "%BUILD_DIR%\service\configs\" >nul 2>&1
)

:: 澶嶅埗璇佷功鏂囦欢
if exist "%SERVICE_DIR%\certs" (
    if not exist "%BUILD_DIR%\service\certs" (
        mkdir "%BUILD_DIR%\service\certs"
    )
    xcopy /s /e /y /q "%SERVICE_DIR%\certs\*" "%BUILD_DIR%\service\certs\" >nul 2>&1
)

:: 澶嶅埗LiveKit閰嶇疆鏂囦欢
if exist "%INFRA_DIR%\docker-compose-livekit.yml" (
    copy /y "%INFRA_DIR%\docker-compose-livekit.yml" "%BUILD_DIR%\"
) else if exist "docker-compose-livekit.yml" (
    copy /y "docker-compose-livekit.yml" "%BUILD_DIR%\"
)

if exist "%INFRA_DIR%\livekit.yaml" (
    copy /y "%INFRA_DIR%\livekit.yaml" "%BUILD_DIR%\"
) else if exist "livekit.yaml" (
    copy /y "livekit.yaml" "%BUILD_DIR%\"
)

:: 澶嶅埗Nginx閰嶇疆鏂囦欢
if exist "%INFRA_DIR%\nginx.conf" (
    copy /y "%INFRA_DIR%\nginx.conf" "%BUILD_DIR%\"
) else if exist "nginx.conf" (
    copy /y "nginx.conf" "%BUILD_DIR%\"
)

:: 澶嶅埗閮ㄧ讲鑴氭湰
if exist "deploy.sh" (
    copy /y "deploy.sh" "%BUILD_DIR%\"
)

if exist "%INFRA_DIR%\deploy-livekit-config.sh" (
    copy /y "%INFRA_DIR%\deploy-livekit-config.sh" "%BUILD_DIR%\"
) else if exist "deploy-livekit-config.sh" (
    copy /y "deploy-livekit-config.sh" "%BUILD_DIR%\"
)

if exist "%INFRA_DIR%\check-deployment.sh" (
    copy /y "%INFRA_DIR%\check-deployment.sh" "%BUILD_DIR%\"
) else if exist "check-deployment.sh" (
    copy /y "check-deployment.sh" "%BUILD_DIR%\"
)

:: Android鏋勫缓宸插垎绂诲埌鐙珛鑴氭湰
echo.
echo [6/6] Android build separated to build-android.bat script
if exist "%ANDROID_DIR%" (
    echo Android project found. Use 'build-android.bat' to build Android app separately.
) else (
    echo Android directory not found, skipping Android build
)

:: 鍒涘缓閮ㄧ讲鍖呭帇缂╂枃浠?
echo.
echo [7/7] Creating deployment package...
cd "%BUILD_DIR%"
if exist "rvcs-%TIMESTAMP%.zip" (
    del /f "rvcs-%TIMESTAMP%.zip"
)
powershell -Command "Compress-Archive -Path * -DestinationPath rvcs-%TIMESTAMP%.zip"
cd ..

:: 娓呯悊Service涓殑涓存椂web鏂囦欢
echo.
echo Cleaning temporary files...
if exist "%SERVICE_DIR%\web" (
    rd /s /q "%SERVICE_DIR%\web"
)

:: 瀹屾垚
echo.
echo ========================================
echo Build completed successfully!
echo ========================================
echo.
echo Output directory: %BUILD_DIR%\
echo   - service/ : Integrated RVCS server with web frontend
echo   - rvcs-%TIMESTAMP%.zip : Deployment package
echo   - Android app: Use separate 'build-android.bat' script
echo.
echo To deploy the integrated service:
echo   1. Copy dist/service to your server
echo   2. Configure dist/service/configs/config.yaml
echo   3. Run: rvcs-server.exe
echo   4. Access web interface at: https://localhost:28443
echo.
echo To build Android app separately: run build-android.bat
echo.

endlocal


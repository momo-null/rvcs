@echo off
:: ========================================
:: RVCS Service Build Script (Windows)
:: Remote Vision & Control System
:: Compatible with Go 1.25+
:: ========================================

echo ========================================
echo RVCS Service Build Script
echo Remote Vision & Control System
echo ========================================
echo.

echo Checking dist directory...
if not exist ..\dist mkdir ..\dist

echo Downloading Go dependencies...
set GOSUMDB=off
set GONOSUMDB=github.com/bytedance/sonic/loader
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

echo.
echo ========================================
echo Build completed successfully!
echo ========================================
echo.
echo Output: ..\dist\rvcs-server.exe
if exist ..\dist\rvcs-server.exe dir ..\dist\rvcs-server.exe
echo.

endlocal

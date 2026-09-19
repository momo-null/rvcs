#!/bin/bash

# ========================================
# RVCS Build Script (Linux/Mac)
# Remote Vision & Control System
# ========================================

set -e  # Exit on error

echo "========================================"
echo "RVCS Build Script"
echo "Remote Vision & Control System"
echo "========================================"
echo ""

# Set variables
BUILD_DIR="dist"
SERVICE_DIR="service"
WEB_DIR="web"
ANDROID_DIR="android"
INFRA_DIR="infra"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Clean old build directory
echo "[1/6] Cleaning build directory..."
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# Build Web frontend
echo ""
echo "[2/6] Building Web frontend..."
cd "$WEB_DIR"

echo "Installing npm dependencies..."
npm install
npm run build
cd ..

# Integrate web files into Service
echo ""
echo "[3/6] Integrating web files into service..."
mkdir -p "$SERVICE_DIR/web"
cp -r "$WEB_DIR/dist/"* "$SERVICE_DIR/web/"

# Build Service backend with integrated web frontend
echo ""
echo "[4/6] Building Service backend with integrated web frontend..."
cd "$SERVICE_DIR"
echo "Setting Go environment..."
export GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,direct
export GOSUMDB=off
echo "Downloading Go dependencies..."
go mod tidy
if [ $? -ne 0 ]; then
    echo "ERROR: Go mod tidy failed!"
    echo "Trying go mod download instead..."
    go mod download
    go mod tidy
    if [ $? -ne 0 ]; then
        echo "ERROR: Go mod tidy still failed!"
        cd ..
        exit 1
    fi
fi
# Build integrated service with web frontend
go build -o rvcs-server -ldflags="-s -w" ./cmd/server
cd ..

# Copy Service to dist directory
echo ""
echo "[5/6] Copying integrated service files to dist..."
mkdir -p "$BUILD_DIR/service"

#set GOSUMDB=off
#set GONOSUMDB=github.com/bytedance/sonic/loader
# Copy executable
cp "$SERVICE_DIR/rvcs-server" "$BUILD_DIR/service/"

# Copy web static files to dist/service
if [ -d "$SERVICE_DIR/web" ]; then
    cp -r "$SERVICE_DIR/web" "$BUILD_DIR/service/"
fi

# Copy config files
if [ -d "$SERVICE_DIR/configs" ]; then
    cp -r "$SERVICE_DIR/configs" "$BUILD_DIR/service/"
fi

# Copy certificate files
if [ -d "$SERVICE_DIR/certs" ]; then
    cp -r "$SERVICE_DIR/certs" "$BUILD_DIR/service/"
fi

# Copy LiveKit / Docker deployment files (keep dist layout unchanged)
if [ -f "$INFRA_DIR/docker-compose-livekit.yml" ]; then
    cp "$INFRA_DIR/docker-compose-livekit.yml" "$BUILD_DIR/"
elif [ -f "docker-compose-livekit.yml" ]; then
    cp "docker-compose-livekit.yml" "$BUILD_DIR/"
fi

if [ -f "$INFRA_DIR/livekit.yaml" ]; then
    cp "$INFRA_DIR/livekit.yaml" "$BUILD_DIR/"
elif [ -f "livekit.yaml" ]; then
    cp "livekit.yaml" "$BUILD_DIR/"
fi

if [ -f "$INFRA_DIR/nginx.conf" ]; then
    cp "$INFRA_DIR/nginx.conf" "$BUILD_DIR/"
elif [ -f "nginx.conf" ]; then
    cp "nginx.conf" "$BUILD_DIR/"
fi

if [ -f "$INFRA_DIR/turnserver.conf" ]; then
    cp "$INFRA_DIR/turnserver.conf" "$BUILD_DIR/"
elif [ -f "turnserver.conf" ]; then
    cp "turnserver.conf" "$BUILD_DIR/"
fi

# Copy deployment scripts
if [ -f "deploy.sh" ]; then
    cp "deploy.sh" "$BUILD_DIR/"
    chmod +x "$BUILD_DIR/deploy.sh"
fi

if [ -f "$INFRA_DIR/deploy-livekit-config.sh" ]; then
    cp "$INFRA_DIR/deploy-livekit-config.sh" "$BUILD_DIR/"
    chmod +x "$BUILD_DIR/deploy-livekit-config.sh"
elif [ -f "deploy-livekit-config.sh" ]; then
    cp "deploy-livekit-config.sh" "$BUILD_DIR/"
    chmod +x "$BUILD_DIR/deploy-livekit-config.sh"
fi

if [ -f "$INFRA_DIR/check-deployment.sh" ]; then
    cp "$INFRA_DIR/check-deployment.sh" "$BUILD_DIR/"
    chmod +x "$BUILD_DIR/check-deployment.sh"
elif [ -f "check-deployment.sh" ]; then
    cp "check-deployment.sh" "$BUILD_DIR/"
    chmod +x "$BUILD_DIR/check-deployment.sh"
fi

# Make server executable
chmod +x "$BUILD_DIR/service/rvcs-server"

# Android build separated to separate script
echo ""
echo "[6/6] Android build separated to build-android.sh script"
if [ -d "$ANDROID_DIR" ]; then
    echo "Android project found. Use './build-android.sh' to build Android app separately."
else
    echo "Android directory not found, skipping Android build"
fi

# Create deployment package
echo ""
echo "[7/7] Creating deployment package..."
cd "$BUILD_DIR"
tar -czf "../rvcs-$TIMESTAMP.tar.gz" *
cd ..

# Clean temporary files in Service
echo ""
echo "Cleaning temporary files..."
rm -rf "$SERVICE_DIR/web"

# Completion
echo ""
echo "========================================"
echo "Build completed successfully!"
echo "========================================"
echo ""
echo "Output directory: $BUILD_DIR/"
echo "  - service/ : Integrated RVCS server with web frontend"
echo "  - rvcs-$TIMESTAMP.tar.gz : Deployment package"
echo "  - Android app: Use separate './build-android.sh' script"
echo ""
echo "To deploy the integrated service:"
echo "  1. Copy dist/service to your server"
echo "  2. Configure dist/service/configs/config.yaml"
echo "  3. Run: ./rvcs-server"
echo "  4. Access web interface at: https://localhost:28443"
echo ""
echo "To build Android app separately: run ./build-android.sh"
echo ""

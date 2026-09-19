# RVCS

> RVCS — Remote Vision & Control System: a self-hosted platform to register, monitor, and remotely control Android camera devices (Go backend + Vue web console + Android client), with live video/audio streaming over LiveKit.

远程设备视频与控制系统，包含三端：
- `service`：Go 后端（HTTP/WebSocket、设备管理、命令下发、心跳）
- `web`：Vue 管理端（设备管理、远程控制、日志、系统设置）
- `android`：设备端（前台服务、心跳、WebSocket、LiveKit 推流）

> **技术底座**：`Go` · `Vue` · `LiveKit`(WebRTC) · `WebSocket` · `Android`(前台服务 / AccessibilityService) · `JWT` 认证 · 自托管部署。

> **免责声明 / Disclaimer**：本项目仅供技术学习与自有设备使用。使用者须确保在其拥有或已获明确同意的设备上部署，并自行承担全部法律责任；作者不对任何第三方使用负责，亦不提供任何担保。

## 当前能力
- 设备注册与认证（JWT）
- 心跳上报与在线状态
- WebSocket 实时控制（`camera_on` / `camera_off` 等）
- 命令队列兜底（`pending` 拉取 + `ack` 回执）
- LiveKit 视频推流 + 网页端对讲
- 熄屏保活（Foreground Service + WakeLock + WifiLock + Watchdog）

## 快速启动

### 1. 后端
```bash
cd service
go run cmd/server/main.go
```
默认读取 `service/configs/config.yaml`。

### 2. 前端
```bash
cd web
npm install
npm run dev
```

### 3. Android
- 用 Android Studio 打开 `android/`
- 修改 `android/app/src/main/res/values/strings.xml` 中服务地址
- 安装并启动前台服务

## 关键端口/地址（以配置为准）
- 后端 HTTP: `28080`
- 后端 HTTPS: `28443`
- LiveKit（Android 字符串配置）: `ws://<host>:7880`

## 文档
- 架构：`docs/architecture.md`
- 部署：`docs/deployment.md`
- 运维与排障：`docs/operations.md`
- 文档索引：`docs/README.md`

## Overview (English)

RVCS (Remote Vision & Control System) is a self-hostable platform for registering, monitoring, and remotely controlling Android camera devices. It consists of three components: a Go backend (device authentication via JWT, command dispatch, heartbeat, offline detection, LiveKit token issuance), a Vue web console (device management, remote control, logs, settings), and an Android client (a foreground service with WakeLock/WifiLock keep-alive, a WebSocket control channel, and LiveKit video/audio streaming plus walkie-talkie). It is designed for scenarios where you need to view and operate remote Android device cameras from a central dashboard.

## Compliance & Responsible Use

This project includes device-side remote control capabilities, including an Android `AccessibilityService` for key injection. It is intended **only for devices you own or have explicit authorization to manage**. Read [COMPLIANCE.md](COMPLIANCE.md) for the full responsible-use and legal assessment before deploying.

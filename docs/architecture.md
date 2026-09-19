# 架构说明

## 1. 系统组成

### `service`（Go）
- 提供 REST API（Web 管理端、Android 设备端）
- 提供 WebSocket（Web 用户连接、设备连接）
- 维护设备状态、活动日志、命令队列
- 生成 LiveKit token

### `web`（Vue 3 + Element Plus）
- 设备列表与详情
- 远程摄像头开关控制
- LiveKit 播放与对讲
- 日志查看
- 系统设置（自动刷新间隔）

### `android`（Kotlin）
- 前台服务常驻（`ForegroundCameraService`）
- 心跳上报（`HeartbeatManager`）
- WebSocket 控制与命令轮询兜底（`WebSocketClient`）
- LiveKit 推流（`AutoCameraStreamManager` / `ServiceStateManager`）
- 保活组件（WakeLock/WifiLock/Watchdog）

## 2. 关键链路

### 心跳链路
`Android HeartbeatManager -> POST /api/v1/device/heartbeat -> service 更新 last_seen/status/ip`

### 控制链路（实时）
`Web -> POST /api/v1/web/devices/:id/control -> service -> device websocket -> 执行`

### 控制链路（兜底）
`control` 写入命令队列后，设备轮询：
- `GET /api/v1/device/commands/pending`
- 执行后 `POST /api/v1/device/commands/:id/ack`

### 视频链路
`Android -> LiveKit 房间推流`，`Web -> 进入同房间观看/对讲`

## 3. 稳定性设计
- Android 前台服务 + `START_STICKY`
- `PARTIAL_WAKE_LOCK` + `WIFI_MODE_FULL_HIGH_PERF`
- 心跳 watchdog：任务异常/心跳陈旧自动重启
- WebSocket ping + 自动重连
- 后端离线检测（`offline_checker`）

## 4. 数据与状态
- 设备核心状态：`online/offline/error`
- 心跳更新：`status`、`last_seen`、`ip_address`
- 命令状态：`pending` -> `acked`

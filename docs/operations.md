# 运维与排障

## 1. 常用日志

### Android
建议过滤：
```bash
adb logcat -s "HeartbeatManager:*" "CameraService:*" "WebSocketClient:*" "AutoCameraStreamManager:*"
```

### 后端
关注：
- `/api/v1/device/heartbeat`
- `/api/v1/device/commands/pending`
- `/api/v1/device/commands/:id/ack`
- WebSocket connected / disconnected

## 2. 心跳稳定判断
满足以下条件可认为稳定：
1. Android 周期出现 `Heartbeat success: code=200`
2. `CameraService [HB-WD] heartbeat healthy ...`
3. 后端 `last_seen` 持续更新

## 3. 常见问题

### 设备显示 offline 但刚刚有心跳
- 检查 `offline_timeout` 是否过小
- 检查设备本地时间/服务器时间
- 检查心跳是否偶发超时

### 控制命令发送成功但设备未执行
- 看设备 websocket 是否在线
- 看是否有 pending 命令
- 看设备是否持续拉取 pending
- 看 ack 是否回写

### `camera_off` 后前端出现 LiveKit 轨道 warning
这是停止过程中的事件竞态，通常不影响实际停流。可在前端“停止中”状态忽略迟到 `track` 事件。

## 4. 推荐配置
- 心跳间隔：20s
- 命令轮询：15s
- 离线超时：180s（根据网络情况可调）

## 5. 回归测试步骤
1. 熄屏 5 分钟，确认心跳不中断
2. 熄屏状态远程 `camera_on` 成功
3. `camera_off` 后设备端资源释放，心跳仍持续
4. Web 设备列表与详情自动刷新状态/日志

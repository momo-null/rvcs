# 部署说明

## 1. 后端部署

### 配置文件
`service/configs/config.yaml`

重点项：
- `server.http_port` / `server.https_port`
- `database.mode`（当前支持 sqlite-file/mysql）
- `jwt.*`
- `websocket.offline_check_interval` / `websocket.offline_timeout`
- `livekit.*`

### 启动
```bash
cd service
go run cmd/server/main.go
```

### 生产建议
- 使用 HTTPS（`tls.enabled=true`）
- 修改默认 `jwt.secret`
- 使用 MySQL（若需要多实例/高并发）

## 2. 前端部署
```bash
cd web
npm install
npm run build
```

产物目录：`web/dist/`

## 3. Android 部署

### 基础配置
文件：`android/app/src/main/res/values/strings.xml`
- `server_base_url`：后端 HTTPS 地址
- `livekit_server_url`：LiveKit 地址

### 打包
```bash
cd android
./gradlew assembleDebug
```

## 4. 反向代理（可选）
建议将 Web 静态资源与 API 统一到同域名，避免跨域和证书问题。

## 5. 最小验收清单
1. Web 登录成功
2. 设备可注册并在线
3. 心跳日志持续更新
4. `camera_on` / `camera_off` 正常
5. Web 可看到视频并可对讲

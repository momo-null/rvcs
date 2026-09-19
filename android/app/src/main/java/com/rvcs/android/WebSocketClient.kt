package com.rvcs.android

import android.content.Context
import android.os.PowerManager
import android.util.Log
import kotlinx.coroutines.*
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.*
import okhttp3.RequestBody.Companion.toRequestBody
import okio.ByteString
import org.json.JSONObject
import java.util.Collections
import java.util.LinkedHashMap
import java.util.concurrent.TimeUnit
import java.util.UUID

class WebSocketClient(
    private val context: Context,
    private val registrar: DeviceRegistrar,
    private val heartbeatManager: HeartbeatManager,
    private val serviceStateManager: ServiceStateManager? = null
) {
    private val client = SSLUtils.createUnsafeOkHttpClient().newBuilder()
        .pingInterval(20, TimeUnit.SECONDS)
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(20, TimeUnit.SECONDS)
        .writeTimeout(20, TimeUnit.SECONDS)
        .callTimeout(30, TimeUnit.SECONDS)
        .build()
    private var ws: WebSocket? = null
    private var reconnectJob: Job? = null
    private var commandPollingJob: Job? = null
    private var isConnected = false
    private val recentRequestIds = Collections.synchronizedMap(
        object : LinkedHashMap<String, Long>(256, 0.75f, true) {
            override fun removeEldestEntry(eldest: MutableMap.MutableEntry<String, Long>?): Boolean {
                return size > 256
            }
        }
    )

    // 设备信息回调
    var onDeviceInfoUpdate: ((String, String) -> Unit)? = null  // (deviceId, deviceName)
    var onDeviceStatusUpdate: ((String) -> Unit)? = null  // (status)

    companion object {
        private const val TAG = "WebSocketClient"
        private const val COMMAND_POLL_INTERVAL_MS = 15000L
    }

    fun connect() {
        val token = registrar.getAccessToken()
        if (token == null) {
            Log.w(TAG, "No access token, skipping WebSocket connection")
            return
        }

        val server = registrar.getServerUrl()
        val wsUrl = server.replaceFirst("http", "ws") + "/api/v1/device/ws"

        Log.d(TAG, "=== Connecting to WebSocket ===")
        Log.d(TAG, "Server: $server")
        Log.d(TAG, "WebSocket URL: $wsUrl")
        Log.d(TAG, "Token: ${token.take(20)}...")

        val req = Request.Builder()
            .url(wsUrl)
            .addHeader("Authorization", "Bearer $token")
            .build()

        ws = client.newWebSocket(req, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.i(TAG, "WebSocket connected successfully")
                Log.d(TAG, "Response code: ${response.code}")
                Log.d(TAG, "Response headers: ${response.headers}")
                isConnected = true
                reconnectJob?.cancel()
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                Log.i(TAG, "WebSocket message received")
                Log.d(TAG, "Message content: $text")
                handleMessage(text)
            }

            override fun onMessage(webSocket: WebSocket, bytes: ByteString) {
                Log.d(TAG, "WebSocket binary message received, size: ${bytes.size}")
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.w(TAG, "WebSocket closing - code: $code, reason: $reason")
                isConnected = false
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.w(TAG, "WebSocket closed - code: $code, reason: $reason")
                isConnected = false
                scheduleReconnect()
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e(TAG, "WebSocket connection failed", t)
                Log.e(TAG, "Response: ${response?.code} ${response?.message}")
                isConnected = false
                scheduleReconnect()
            }
        })
    }

    private fun handleMessage(text: String) {
        try {
            Log.d(TAG, "=== Handling WebSocket Message ===")
            Log.d(TAG, "Raw message: $text")

            val json = JSONObject(text)
            val type = json.optString("type")

            Log.d(TAG, "Message type: $type")

            when (type) {
                "control" -> {
                    Log.d(TAG, "Processing control command")
                    handleControlCommand(json)
                }
                "device_update" -> {
                    Log.d(TAG, "Processing device update")
                    handleDeviceUpdate(json)
                }
                "status_update" -> {
                    Log.d(TAG, "Processing status update")
                    handleStatusUpdate(json)
                }
                "device_status_changed" -> {
                    Log.d(TAG, "Processing device status changed")
                    handleStatusUpdate(json)
                }
                else -> {
                    Log.w(TAG, "Unknown message type: $type")
                    Log.d(TAG, "Message content: $text")
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error handling message", e)
            Log.e(TAG, "Message content: $text")
        }
    }

    private fun handleControlCommand(json: JSONObject) {
        try {
            Log.d(TAG, "=== Processing Control Command ===")

            // 从data字段中提取命令信息
            val data = json.optJSONObject("data")
            val command = data?.optString("command") ?: json.optString("command", "")
            val requestId = data?.optString("request_id", "") ?: json.optString("request_id", "")
            val resolution = data?.optJSONObject("data")?.optString("resolution") ?: data?.optString("resolution") ?: "720p"

            Log.i(TAG, "Control command received")
            Log.d(TAG, "Command: '$command'")
            Log.d(TAG, "Request ID: '$requestId'")
            Log.d(TAG, "Resolution: '$resolution'")
            Log.d(TAG, "Full JSON: $json")
            Log.d(TAG, "Data object: $data")

            val result = executeControlCommand(command, resolution, requestId, data)

            Log.d(TAG, "Command execution result: $result")
            sendCommandResponse(requestId, command, result)
        } catch (e: Exception) {
            Log.e(TAG, "Error handling control command", e)
        }
    }

    fun executeControlCommand(
        command: String,
        resolution: String? = null,
        requestId: String = "",
        data: JSONObject? = null
    ): Boolean {
        if (requestId.isNotBlank()) {
            val isDup = synchronized(recentRequestIds) {
                val exists = recentRequestIds.containsKey(requestId)
                if (!exists) {
                    recentRequestIds[requestId] = System.currentTimeMillis()
                }
                exists
            }
            if (isDup) {
                Log.w(TAG, "Duplicate command ignored by request_id=$requestId")
                return true
            }
        }

        val result = when (command) {
            "camera_on" -> {
                Log.i(TAG, "Executing camera_on command")
                startCamera(resolution)
            }
            "camera_off" -> {
                Log.i(TAG, "Executing camera_off command")
                stopCamera()
            }
            "reboot" -> {
                Log.i(TAG, "Executing reboot command")
                rebootDevice()
            }
            "set_bitrate" -> setBitrate(data)
            "set_resolution" -> setResolution(data)
            else -> {
                Log.w(TAG, "Unknown command: $command")
                false
            }
        }
        return result
    }

    private fun startCamera(): Boolean {
        return startCamera(null)
    }

    private fun startCamera(resolutionString: String?): Boolean {
        return try {
            Log.i(TAG, "=== Starting Camera and Streaming ===")

            // 检查 ServiceStateManager 是否初始化
            if (serviceStateManager == null) {
                Log.e(TAG, "ServiceStateManager is null, cannot start camera")
                sendCameraStateChanged("error", "Service not initialized")
                return false
            }

            // 先清理旧的推流（如果有）
            val streamManager = serviceStateManager?.getStreamManager()
            if (streamManager != null && streamManager.isStreaming()) {
                Log.w(TAG, "检测到已有推流，先停止...")
                serviceStateManager?.stopStreaming()
                Thread.sleep(500)
            }

            // 解析分辨率参数
            val resolution = resolutionString?.let {
                VideoResolution.fromString(it)
            } ?: VideoResolution.RES_720P
            Log.d(TAG, "Using resolution: ${resolution.displayName}")

            // 设置分辨率
            val resolutionResult = serviceStateManager?.setResolution(resolution) ?: false
            if (!resolutionResult) {
                Log.w(TAG, "Failed to set resolution, using default")
            }

            // 检查摄像头权限
            val context = this.context
            if (android.content.pm.PackageManager.PERMISSION_GRANTED !=
                context.checkSelfPermission(android.Manifest.permission.CAMERA)) {
                Log.e(TAG, "Camera permission not granted")
                sendCameraStateChanged("error", "Camera permission not granted")
                return false
            }

            // 直接启动推流（SDK 自动管理摄像头）
            Log.d(TAG, "Starting streaming (SDK auto-manages camera)...")

            val cameraResult = serviceStateManager?.startCamera() ?: false
            Log.d(TAG, "Camera start result: $cameraResult")

            if (!cameraResult) {
                Log.e(TAG, "Failed to start camera")
                sendCameraStateChanged("error", "Failed to start camera")
                return false
            }

            // 等待摄像头完全启动
            Log.d(TAG, "Waiting for camera to fully start...")
            Thread.sleep(1000)

            // 启动推流（使用 LiveKit）
            Log.d(TAG, "Starting LiveKit streaming...")
            val streamResult = serviceStateManager?.startStreaming() ?: false
            Log.d(TAG, "Stream start result: $streamResult")

            if (!streamResult) {
                Log.e(TAG, "Failed to start streaming")
                // 推流失败，停止摄像头
                serviceStateManager?.stopCamera()
                sendCameraStateChanged("error", "Failed to start streaming")
                return false
            }

            // 更新心跳状态
            sendCameraStateChanged("on", "Camera and streaming started successfully")
            Log.i(TAG, "✅ Camera and streaming started successfully")
            true
        } catch (e: SecurityException) {
            Log.e(TAG, "Security exception when starting camera", e)
            sendCameraStateChanged("error", "Camera permission denied")
            false
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start camera", e)
            sendCameraStateChanged("error", "Failed to start camera: ${e.message}")
            false
        }
    }

    private fun stopCamera(): Boolean {
        return try {
            Log.i(TAG, "=== Stopping Camera and Streaming ===")

            // 检查 ServiceStateManager 是否初始化
            if (serviceStateManager == null) {
                Log.e(TAG, "ServiceStateManager is null, cannot stop camera")
                sendCameraStateChanged("error", "Service not initialized")
                return false
            }

            // 停止推流（会自动停止摄像头）
            Log.d(TAG, "Stopping streaming...")
            val result = serviceStateManager?.stopStreaming() ?: false
            Log.d(TAG, "Streaming stop result: $result")

            if (result) {
                // 更新心跳状态
                sendCameraStateChanged("off", "Camera and streaming stopped successfully")
                Log.i(TAG, "✅ Camera and streaming stopped successfully")
            } else {
                Log.e(TAG, "Failed to stop camera")
                sendCameraStateChanged("error", "Failed to stop camera")
            }

            result
        } catch (e: Exception) {
            Log.e(TAG, "Failed to stop camera", e)
            sendCameraStateChanged("error", "Failed to stop camera: ${e.message}")
            false
        }
    }

    private fun rebootDevice(): Boolean {
        return try {
            Log.d(TAG, "Rebooting device")

            // 先停止服务和摄像头
            serviceStateManager?.stopCamera()

            // 重启设备
            try {
                val powerManager = context.getSystemService(Context.POWER_SERVICE) as PowerManager
                powerManager.reboot("Remote reboot command")
                true
            } catch (e: Exception) {
                Log.e(TAG, "Failed to reboot device", e)
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to reboot device", e)
            false
        }
    }

    /**
     * 设置码率
     */
    private fun setBitrate(data: JSONObject?): Boolean {
        return try {
            Log.d(TAG, "Setting bitrate")
            val targetKbps = data?.optInt("target_kbps") ?: 1200
            val minKbps = data?.optInt("min_kbps")
            val maxKbps = data?.optInt("max_kbps")
            val adaptive = data?.optBoolean("adaptive") ?: true

            Log.d(TAG, "Bitrate settings - Target: ${targetKbps}kbps, Min: ${minKbps}, Max: ${maxKbps}, Adaptive: $adaptive")

            // 设置码率
            val result = serviceStateManager?.setBitrate(targetKbps, minKbps, maxKbps) ?: false

            // 启用/禁用自适应码率
            serviceStateManager?.setAdaptiveBitrate(adaptive)

            if (result) {
                Log.i(TAG, "Bitrate set successfully: ${targetKbps}kbps (adaptive: $adaptive)")
                true
            } else {
                Log.e(TAG, "Failed to set bitrate")
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to set bitrate", e)
            false
        }
    }

    /**
     * 设置分辨率
     */
    private fun setResolution(data: JSONObject?): Boolean {
        return try {
            Log.d(TAG, "Setting resolution")
            val resolutionString = data?.optString("resolution") ?: "720p"

            Log.d(TAG, "Resolution request: $resolutionString")

            // 解析分辨率
            val resolution = VideoResolution.fromString(resolutionString)

            // 设置分辨率
            val result = serviceStateManager?.setResolution(resolution) ?: false

            if (result) {
                Log.i(TAG, "Resolution set successfully: ${resolution.displayName}")
                true
            } else {
                Log.e(TAG, "Failed to set resolution")
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to set resolution", e)
            false
        }
    }

    private fun sendCommandResponse(requestId: String, command: String, success: Boolean) {
        try {
            val now = System.currentTimeMillis()
            val msgId = UUID.randomUUID().toString()
            val data = JSONObject().apply {
                put("request_id", requestId)
                put("command", command)
                put("success", success)
                put("timestamp", now)
            }
            val response = JSONObject()
            response.put("msg_id", msgId)
            response.put("type", "command_response")
            response.put("command", command)
            response.put("status", if (success) "success" else "failed")
            response.put("timestamp", now)
            response.put("data", data)

            val responseStr = response.toString()
            Log.d(TAG, "Sending command response: $responseStr")

            if (ws?.send(responseStr) == true) {
                Log.d(TAG, "✅ Command response sent successfully")
            } else {
                Log.e(TAG, "❌ Failed to send command response: WebSocket is null")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send command response", e)
        }
    }

    private fun sendCameraStateChanged(state: String, message: String) {
        try {
            val now = System.currentTimeMillis()
            val msgId = UUID.randomUUID().toString()
            val data = JSONObject().apply {
                put("state", state)
                put("message", message)
                put("timestamp", now)
            }
            val notification = JSONObject()
            notification.put("msg_id", msgId)
            notification.put("type", "camera_state_changed")
            notification.put("state", state)
            notification.put("status", if (state == "error") "error" else "ok")
            notification.put("timestamp", now)
            notification.put("data", data)

            val notificationStr = notification.toString()
            Log.d(TAG, "Sending camera state changed: $notificationStr")

            if (ws?.send(notificationStr) == true) {
                Log.d(TAG, "✅ Camera state changed notification sent successfully")
            } else {
                Log.e(TAG, "❌ Failed to send camera state changed notification: WebSocket is null")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send camera state changed notification", e)
        }
    }

    private fun handleDeviceUpdate(json: JSONObject) {
        try {
            val deviceData = json.optJSONObject("data")
            Log.d(TAG, "Device update received: $deviceData")

            // 更新本地设备信息
            if (deviceData != null) {
                val deviceId = deviceData.optString("id")
                val deviceName = deviceData.optString("name")
                val manufacturer = deviceData.optString("manufacturer")
                val androidVersion = deviceData.optString("android_version")
                val appVersion = deviceData.optString("app_version")

                // 更新 DeviceRegistrar 中的信息
                if (deviceId.isNotEmpty() && deviceId != registrar.getDeviceId()) {
                    Log.d(TAG, "Updating device ID: $deviceId")
                    registrar.setDeviceId(deviceId)
                }

                Log.d(TAG, "Device info updated - ID: $deviceId, Name: $deviceName, Manufacturer: $manufacturer")

                // 触发回调通知
                onDeviceInfoUpdate?.invoke(deviceId, deviceName)
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error handling device update", e)
        }
    }

    private fun handleStatusUpdate(json: JSONObject) {
        try {
            val status = json.optString("status")
            val deviceData = json.optJSONObject("data")

            Log.d(TAG, "Status update received: $status")

            // 更新设备状态
            if (deviceData != null) {
                val deviceId = deviceData.optString("id")
                val newStatus = deviceData.optString("status")

                // 如果更新的是当前设备
                if (deviceId == registrar.getDeviceId()) {
                    Log.d(TAG, "Device status updated to: $newStatus")

                    // 根据状态调整服务行为
                    when (newStatus) {
                        "offline" -> {
                            // 设备被标记为离线，可能需要停止推流
                            serviceStateManager?.stopCamera()
                            Log.w(TAG, "Device marked as offline, stopping camera")
                        }
                        "error" -> {
                            // 设备处于错误状态
                            Log.e(TAG, "Device in error state")
                        }
                        "online" -> {
                            Log.d(TAG, "Device is online")
                        }
                    }

                    // 触发回调通知
                    onDeviceStatusUpdate?.invoke(newStatus)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error handling status update", e)
        }
    }

    private fun scheduleReconnect() {
        reconnectJob?.cancel()
        reconnectJob = CoroutineScope(Dispatchers.IO).launch {
            delay(5000)
            Log.d(TAG, "Attempting to reconnect...")
            connect()
        }
    }

    fun startCommandPolling(intervalMs: Long = COMMAND_POLL_INTERVAL_MS) {
        if (commandPollingJob?.isActive == true) return
        commandPollingJob = CoroutineScope(Dispatchers.IO).launch {
            while (isActive) {
                try {
                    pollPendingCommands()
                } catch (e: Exception) {
                    Log.w(TAG, "Command polling error: ${e.message}")
                }
                delay(intervalMs)
            }
        }
        Log.i(TAG, "Command polling started, interval=${intervalMs}ms")
    }

    fun stopCommandPolling() {
        commandPollingJob?.cancel()
        commandPollingJob = null
        Log.i(TAG, "Command polling stopped")
    }

    private fun pollPendingCommands() {
        val token = registrar.getAccessToken() ?: return
        val server = registrar.getServerUrl()
        val url = "$server/api/v1/device/commands/pending?limit=10"

        val req = Request.Builder()
            .url(url)
            .get()
            .addHeader("Authorization", "Bearer $token")
            .build()

        client.newCall(req).execute().use { resp ->
            if (!resp.isSuccessful) {
                Log.w(TAG, "Pull pending commands failed: ${resp.code}")
                return
            }

            val text = resp.body?.string().orEmpty()
            if (text.isBlank()) return

            val root = JSONObject(text)
            val data = root.optJSONObject("data") ?: return
            val items = data.optJSONArray("items") ?: return
            if (items.length() == 0) return

            Log.i(TAG, "Pulled ${items.length()} pending command(s)")
            for (i in 0 until items.length()) {
                val item = items.optJSONObject(i) ?: continue
                val cmdId = item.optLong("id", 0L)
                val payload = item.optJSONObject("payload")
                val command = item.optString("command").ifBlank { payload?.optString("command") ?: "" }
                val requestId = item.optString("request_id").ifBlank { payload?.optString("request_id") ?: "" }
                val resolution = payload?.optJSONObject("data")?.optString("resolution")
                    ?: payload?.optString("resolution")

                if (command.isBlank()) {
                    ackCommand(cmdId, false, "empty command")
                    continue
                }

                val ok = executeControlCommand(command, resolution, requestId)
                ackCommand(cmdId, ok, if (ok) "ok" else "execute failed")
            }
        }
    }

    private fun ackCommand(commandId: Long, success: Boolean, result: String) {
        if (commandId <= 0) return
        val token = registrar.getAccessToken() ?: return
        val server = registrar.getServerUrl()
        val url = "$server/api/v1/device/commands/$commandId/ack"
        val bodyJson = JSONObject().apply {
            put("success", success)
            put("result", result)
        }
        val body = bodyJson.toString().toRequestBody("application/json".toMediaTypeOrNull())
        val req = Request.Builder()
            .url(url)
            .post(body)
            .addHeader("Authorization", "Bearer $token")
            .build()
        try {
            client.newCall(req).execute().use { resp ->
                if (!resp.isSuccessful) {
                    Log.w(TAG, "Ack command failed: id=$commandId code=${resp.code}")
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Ack command error: id=$commandId ${e.message}")
        }
    }

    fun disconnect() {
        reconnectJob?.cancel()
        stopCommandPolling()
        ws?.close(1000, "client_close")
        ws = null
        isConnected = false
    }

    fun isConnected(): Boolean = isConnected
}

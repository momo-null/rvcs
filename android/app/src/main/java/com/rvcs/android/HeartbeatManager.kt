package com.rvcs.android

import android.content.Context
import android.content.SharedPreferences
import android.os.BatteryManager
import android.util.Log
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import java.util.concurrent.atomic.AtomicLong
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import java.util.concurrent.TimeUnit

class HeartbeatManager(private val context: Context, private val registrar: DeviceRegistrar) {
    private val client = SSLUtils.createUnsafeOkHttpClient().newBuilder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(10, TimeUnit.SECONDS)
        .writeTimeout(10, TimeUnit.SECONDS)
        .callTimeout(15, TimeUnit.SECONDS)
        .build()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private var job: Job? = null
    private var isStreaming = false
    private val lastSuccessAtMs = AtomicLong(0L)
    private val lastLoopTickAtMs = AtomicLong(0L)
    private val lastAlertAtMs = AtomicLong(0L)
    @Volatile
    private var lastAlertReason: String = ""
    private val diagPrefs: SharedPreferences =
        context.getSharedPreferences(DIAG_PREFS, Context.MODE_PRIVATE)

    companion object {
        private const val TAG = "HeartbeatManager"
        private const val MAX_BACKOFF_MS = 5 * 60 * 1000L
        private const val DIAG_PREFS = "rvcs_diag_prefs"
        private const val KEY_LAST_HEARTBEAT_OK_MS = "last_heartbeat_ok_ms"
        private const val KEY_LAST_HEARTBEAT_LOOP_TICK_MS = "last_heartbeat_loop_tick_ms"
        private const val KEY_LAST_HEARTBEAT_FAIL_MS = "last_heartbeat_fail_ms"
        private const val KEY_LAST_HEARTBEAT_FAIL_REASON = "last_heartbeat_fail_reason"
        private const val ALERT_COOLDOWN_MS = 10 * 60 * 1000L
    }

    fun start(intervalMs: Long = 30000L) {
        if (job?.isActive == true) {
            Log.d(TAG, "Heartbeat already running")
            return
        }

        job = scope.launch {
            var currentDelay = intervalMs
            var consecutiveFailures = 0

            Log.i(TAG, "Heartbeat started, interval=${intervalMs}ms")
            while (isActive) {
                val now = System.currentTimeMillis()
                lastLoopTickAtMs.set(now)
                diagPrefs.edit().putLong(KEY_LAST_HEARTBEAT_LOOP_TICK_MS, now).apply()

                val ok = try {
                    sendHeartbeat()
                } catch (ce: CancellationException) {
                    throw ce
                } catch (e: Exception) {
                    recordFailure("HB_LOOP_CRASH:${e.javaClass.simpleName}")
                    Log.e(TAG, "Heartbeat loop crashed once, keep running", e)
                    false
                }
                if (ok) {
                    consecutiveFailures = 0
                    currentDelay = intervalMs
                } else {
                    consecutiveFailures += 1
                    currentDelay = (intervalMs * (1L shl minOf(consecutiveFailures, 4))).coerceAtMost(MAX_BACKOFF_MS)
                    Log.w(TAG, "Heartbeat failed, retry in ${currentDelay}ms (failures=$consecutiveFailures)")
                }
                delay(currentDelay)
            }
        }
    }

    fun stop() {
        job?.cancel()
        job = null
        Log.i(TAG, "Heartbeat stopped")
    }

    fun isRunning(): Boolean = job?.isActive == true

    fun getLastSuccessAtMs(): Long = lastSuccessAtMs.get()
    fun getLastLoopTickAtMs(): Long = lastLoopTickAtMs.get()

    fun setStreamingState(streaming: Boolean) {
        isStreaming = streaming
        Log.d(TAG, "Streaming state updated: $streaming")
    }

    fun isStreaming(): Boolean = isStreaming

    private suspend fun sendHeartbeat(): Boolean {
        return try {
            val deviceId = registrar.getDeviceId()
            val token = registrar.getAccessToken()

            if (deviceId.isNullOrEmpty() || token.isNullOrEmpty()) {
                Log.w(TAG, "Device/token unavailable, heartbeat skipped")
                false
            } else {
                val server = registrar.getServerUrl()
                val url = "$server/api/v1/device/heartbeat"

                val payload = JSONObject().apply {
                    put("status", "online")
                    put("health", JSONObject().apply {
                        put("battery", getBatteryLevel())
                        put("memory_usage", getMemoryUsage())
                        put("streaming", isStreaming)
                    })
                }

                val body = payload.toString().toRequestBody("application/json".toMediaTypeOrNull())
                val req = Request.Builder()
                    .url(url)
                    .post(body)
                    .addHeader("Authorization", "Bearer $token")
                    .build()

                Log.d(TAG, "Sending heartbeat request...")
                client.newCall(req).execute().use { resp ->
                    when {
                        resp.isSuccessful -> {
                            val text = resp.body?.string().orEmpty()
                            Log.d(TAG, "Heartbeat success: code=${resp.code}")
                            val okAt = System.currentTimeMillis()
                            lastSuccessAtMs.set(okAt)
                            diagPrefs.edit()
                                .putLong(KEY_LAST_HEARTBEAT_OK_MS, okAt)
                                .putString(KEY_LAST_HEARTBEAT_FAIL_REASON, "")
                                .apply()
                            if (text.isNotEmpty()) {
                                try {
                                    val responseJson = JSONObject(text)
                                    if (responseJson.has("data")) {
                                        val data = responseJson.getJSONObject("data")
                                        if (data.has("config_update")) {
                                            Log.d(TAG, "Heartbeat config update received")
                                        }
                                    }
                                } catch (e: Exception) {
                                    Log.w(TAG, "Failed to parse heartbeat response: ${e.message}")
                                }
                            }
                            true
                        }

                        resp.code == 401 -> {
                            Log.w(TAG, "Heartbeat got 401, try refresh token")
                            recordFailure("HB_HTTP_401")
                            val refreshed = registrar.refreshToken()
                            if (refreshed) {
                                Log.i(TAG, "Token refresh success, will retry heartbeat")
                            } else {
                                Log.e(TAG, "Token refresh failed, keep heartbeat loop alive")
                            }
                            false
                        }

                        else -> {
                            val err = resp.body?.string().orEmpty()
                            Log.e(TAG, "Heartbeat failed: ${resp.code} $err")
                            recordFailure("HB_HTTP_${resp.code}")
                            false
                        }
                    }
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Heartbeat error", e)
            recordFailure("HB_SEND_EXCEPTION:${e.javaClass.simpleName}")
            false
        }
    }

    private fun recordFailure(reason: String) {
        val now = System.currentTimeMillis()
        diagPrefs.edit()
            .putLong(KEY_LAST_HEARTBEAT_FAIL_MS, now)
            .putString(KEY_LAST_HEARTBEAT_FAIL_REASON, reason)
            .apply()
        reportHeartbeatAlert(reason)
    }

    fun reportHeartbeatAlert(reason: String) {
        val now = System.currentTimeMillis()
        val previousAt = lastAlertAtMs.get()
        val previousReason = lastAlertReason
        if (previousReason == reason && now - previousAt < ALERT_COOLDOWN_MS) {
            Log.d(TAG, "Skip heartbeat alert by cooldown, reason=$reason")
            return
        }

        val token = registrar.getAccessToken()
        if (token.isNullOrBlank()) {
            Log.w(TAG, "Skip heartbeat alert: token unavailable, reason=$reason")
            return
        }
        val deviceId = registrar.getDeviceId().orEmpty()
        val server = registrar.getServerUrl()
        val url = "$server/api/v1/device/alerts"
        val payload = JSONObject().apply {
            put("event_type", "heartbeat_issue")
            put("level", "warning")
            put("message", "Heartbeat anomaly detected: $reason")
            put(
                "data",
                JSONObject().apply {
                    put("reason", reason)
                    put("device_id", deviceId)
                    put("last_heartbeat_ok_ms", getLastSuccessAtMs())
                    put("last_heartbeat_loop_tick_ms", getLastLoopTickAtMs())
                    put("timestamp", now)
                }
            )
        }
        val body = payload.toString().toRequestBody("application/json".toMediaTypeOrNull())
        val req = Request.Builder()
            .url(url)
            .post(body)
            .addHeader("Authorization", "Bearer $token")
            .build()
        try {
            client.newCall(req).execute().use { resp ->
                if (resp.isSuccessful) {
                    lastAlertAtMs.set(now)
                    lastAlertReason = reason
                    Log.w(TAG, "Heartbeat alert reported: reason=$reason code=${resp.code}")
                } else {
                    Log.w(TAG, "Heartbeat alert report failed: code=${resp.code} reason=$reason")
                }
            }
        } catch (e: Exception) {
            Log.w(TAG, "Heartbeat alert report error: reason=$reason msg=${e.message}")
        }
    }

    private fun getBatteryLevel(): Int {
        return try {
            val batteryManager = context.getSystemService(Context.BATTERY_SERVICE) as? BatteryManager
            if (batteryManager != null) {
                batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY)
            } else {
                val intentFilter = android.content.IntentFilter(android.content.Intent.ACTION_BATTERY_CHANGED)
                val batteryStatus = context.registerReceiver(null, intentFilter)
                val level = batteryStatus?.getIntExtra(BatteryManager.EXTRA_LEVEL, -1) ?: -1
                val scale = batteryStatus?.getIntExtra(BatteryManager.EXTRA_SCALE, -1) ?: -1
                if (level != -1 && scale > 0) (level * 100) / scale else -1
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error getting battery level: ${e.message}")
            -1
        }
    }

    private fun getMemoryUsage(): Int {
        return try {
            val runtime = Runtime.getRuntime()
            val usedMemory = runtime.totalMemory() - runtime.freeMemory()
            val maxMemory = runtime.maxMemory()
            val appPercent = ((usedMemory * 100) / maxMemory).toInt()

            val activityManager = context.getSystemService(Context.ACTIVITY_SERVICE) as? android.app.ActivityManager
            val memoryInfo = android.app.ActivityManager.MemoryInfo()
            activityManager?.getMemoryInfo(memoryInfo)

            val totalMem = memoryInfo.totalMem
            val availMem = memoryInfo.availMem
            val systemPercent = if (totalMem > 0) (((totalMem - availMem) * 100) / totalMem).toInt() else appPercent

            (appPercent + systemPercent) / 2
        } catch (e: Exception) {
            Log.e(TAG, "Error getting memory usage: ${e.message}")
            val runtime = Runtime.getRuntime()
            val usedMemory = runtime.totalMemory() - runtime.freeMemory()
            val maxMemory = runtime.maxMemory()
            ((usedMemory * 100) / maxMemory).toInt()
        }
    }
}

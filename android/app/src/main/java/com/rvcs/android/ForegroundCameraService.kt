package com.rvcs.android

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

class ForegroundCameraService : Service() {
    private val serviceScope = CoroutineScope(Dispatchers.Default + Job())
    private var deviceRegistrar: DeviceRegistrar? = null
    private var heartbeatManager: HeartbeatManager? = null
    private var webSocketClient: WebSocketClient? = null
    private var serviceStateManager: ServiceStateManager? = null
    private var serviceWakeLockHelper: WakeLockHelper? = null
    private var serviceWifiLockHelper: WifiLockHelper? = null
    private var batteryOptHelper: BatteryOptimizationHelper? = null
    private var heartbeatWatchdogJob: Job? = null
    private var isRunning = false
    private var heartbeatStaleStrike = 0

    companion object {
        private const val TAG = "CameraService"
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL_ID = "rvcs_service_channel"
        const val ACTION_START = "ACTION_START"
        const val ACTION_STOP = "ACTION_STOP"
        const val ACTION_UPDATE_NOTIFICATION = "ACTION_UPDATE_NOTIFICATION"
        private const val HEARTBEAT_INTERVAL_MS = 20000L
        private const val HEARTBEAT_WATCHDOG_INTERVAL_MS = 60000L
        private const val HEARTBEAT_STALE_THRESHOLD_MS = 120000L
        private const val HEARTBEAT_LOOP_STALE_THRESHOLD_MS = 180000L
    }

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()

        deviceRegistrar = DeviceRegistrar(applicationContext)
        serviceWakeLockHelper = WakeLockHelper(applicationContext)
        serviceWifiLockHelper = WifiLockHelper(applicationContext)
        batteryOptHelper = BatteryOptimizationHelper(applicationContext)
        batteryOptHelper?.logOptimizationStatus()

        Log.d(TAG, "Service created")
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_START -> startServiceInternal()
            ACTION_STOP -> stopServiceInternal()
            ACTION_UPDATE_NOTIFICATION -> updateNotification(intent.getStringExtra("status") ?: "Running")
        }

        return START_STICKY
    }

    private fun startServiceInternal() {
        Log.d(TAG, "=== startService called ===")
        if (isRunning) {
            Log.d(TAG, "Service already running, skipping start")
            return
        }

        Log.i(TAG, "Starting camera service...")
        isRunning = true
        startForegroundCompat()
        updateNotification("Initializing...")
        serviceWakeLockHelper?.startKeepAwake()
        serviceWifiLockHelper?.acquire()
        Log.d(TAG, "Service-level wakelock acquired")

        serviceScope.launch {
            try {
                val isRegistered = deviceRegistrar?.isRegistered()
                Log.d(TAG, "isRegistered: $isRegistered")

                if (isRegistered == true) {
                    val deviceId = deviceRegistrar?.getDeviceId()
                    val token = deviceRegistrar?.getAccessToken()
                    updateNotification("Registered: $deviceId")

                    val liveKitServerUrl = applicationContext.getString(R.string.livekit_server_url)
                    serviceStateManager = ServiceStateManager(
                        context = applicationContext,
                        registrar = deviceRegistrar!!,
                        liveKitServerUrl = liveKitServerUrl
                    )

                    val initialized = if (deviceId != null && token != null) {
                        serviceStateManager?.initialize(deviceId, token)
                    } else {
                        Log.w(TAG, "Device ID or token is null, initializing without camera")
                        serviceStateManager?.initializeWithoutCamera(token!!)
                    }

                    if (initialized == true) {
                        Log.i(TAG, "ServiceStateManager initialized successfully")
                    } else {
                        Log.w(TAG, "ServiceStateManager initialization failed, continuing anyway")
                    }

                    heartbeatManager = HeartbeatManager(applicationContext, deviceRegistrar!!)
                    heartbeatManager?.start(HEARTBEAT_INTERVAL_MS)
                    Log.i(TAG, "Heartbeat started successfully")
                    startHeartbeatWatchdog()

                    webSocketClient = WebSocketClient(
                        applicationContext,
                        deviceRegistrar!!,
                        heartbeatManager!!,
                        serviceStateManager
                    )
                    webSocketClient?.connect()
                    webSocketClient?.startCommandPolling()

                    updateNotification("Service running - Online")
                    Log.i(TAG, "Service started")
                } else {
                    Log.e(TAG, "Device not registered")
                    updateNotification("Not registered")
                    stopSelf()
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error starting service", e)
                updateNotification("Error: ${e.message}")
                stopSelf()
            }
        }
    }

    private fun stopServiceInternal() {
        Log.d(TAG, "Stopping service")

        serviceStateManager?.cleanup()
        serviceStateManager = null

        webSocketClient?.disconnect()
        webSocketClient = null

        heartbeatManager?.stop()
        heartbeatManager = null
        heartbeatWatchdogJob?.cancel()
        heartbeatWatchdogJob = null

        serviceWakeLockHelper?.stopKeepAwake()
        serviceWifiLockHelper?.release()

        isRunning = false
        stopForeground(STOP_FOREGROUND_REMOVE)
        stopSelf()
    }

    private fun startForegroundCompat() {
        val notification: Notification = createNotification("Starting...")
        startForeground(NOTIFICATION_ID, notification)
    }

    private fun createNotification(status: String): Notification {
        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle(getString(R.string.service_name))
            .setContentText(status)
            .setSmallIcon(android.R.drawable.ic_menu_camera)
            .setOngoing(true)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setForegroundServiceBehavior(NotificationCompat.FOREGROUND_SERVICE_IMMEDIATE)
            .build()
    }

    private fun updateNotification(status: String) {
        val notification = createNotification(status)
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification)
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "RVCS Service Channel",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for RVCS camera device service"
                setShowBadge(false)
                enableVibration(false)
                setSound(null, null)
            }

            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        super.onDestroy()
        Log.d(TAG, "Service destroyed")

        serviceStateManager?.cleanup()
        serviceStateManager = null

        webSocketClient?.disconnect()
        heartbeatManager?.stop()
        heartbeatWatchdogJob?.cancel()
        heartbeatWatchdogJob = null
        serviceWakeLockHelper?.cleanup()
        serviceWifiLockHelper?.cleanup()
        serviceScope.coroutineContext[Job]?.cancel()

        deviceRegistrar = null
        heartbeatManager = null
        webSocketClient = null
        serviceWakeLockHelper = null
        serviceWifiLockHelper = null
        isRunning = false
    }

    private fun startHeartbeatWatchdog() {
        heartbeatWatchdogJob?.cancel()
        heartbeatWatchdogJob = serviceScope.launch {
            while (isActive && isRunning) {
                delay(HEARTBEAT_WATCHDOG_INTERVAL_MS)

                val manager = heartbeatManager
                if (manager == null) {
                    Log.w(TAG, "[HB-WD] HeartbeatManager is null, re-create and start")
                    val registrar = deviceRegistrar
                    if (registrar != null) {
                        heartbeatManager = HeartbeatManager(applicationContext, registrar).also {
                            it.start(HEARTBEAT_INTERVAL_MS)
                        }
                    }
                    continue
                }

                if (!manager.isRunning()) {
                    Log.w(TAG, "[HB-WD] Heartbeat job stopped unexpectedly, restarting")
                    manager.start(HEARTBEAT_INTERVAL_MS)
                    heartbeatStaleStrike = 0
                    continue
                }

                val lastLoopTick = manager.getLastLoopTickAtMs()
                if (lastLoopTick > 0L) {
                    val loopStaleMs = System.currentTimeMillis() - lastLoopTick
                    if (loopStaleMs > HEARTBEAT_LOOP_STALE_THRESHOLD_MS) {
                        heartbeatStaleStrike += 1
                        Log.w(
                            TAG,
                            "[HB-WD] Heartbeat loop tick stale ${loopStaleMs}ms (strike=$heartbeatStaleStrike), force restart heartbeat"
                        )
                        manager.reportHeartbeatAlert("HB_STUCK_NO_TICK")
                        manager.stop()
                        manager.start(HEARTBEAT_INTERVAL_MS)

                        if (heartbeatStaleStrike >= 2) {
                            val ws = webSocketClient
                            if (ws != null) {
                                Log.w(TAG, "[HB-WD] Loop stale persists, rebuild websocket stack")
                                ws.disconnect()
                                ws.connect()
                                ws.startCommandPolling()
                            }
                            heartbeatStaleStrike = 0
                        }
                        continue
                    }
                }

                val lastOk = manager.getLastSuccessAtMs()
                if (lastOk > 0L) {
                    val staleMs = System.currentTimeMillis() - lastOk
                    if (staleMs > HEARTBEAT_STALE_THRESHOLD_MS) {
                        heartbeatStaleStrike += 1
                        Log.w(TAG, "[HB-WD] Last heartbeat success is stale (${staleMs}ms), force restart heartbeat job")
                        manager.reportHeartbeatAlert("HB_NO_SUCCESS_STALE")
                        manager.stop()
                        manager.start(HEARTBEAT_INTERVAL_MS)

                        val ws = webSocketClient
                        if (ws == null) {
                            Log.w(TAG, "[HB-WD] WebSocketClient is null, recreating")
                            val registrar = deviceRegistrar
                            if (registrar != null) {
                                webSocketClient = WebSocketClient(
                                    applicationContext,
                                    registrar,
                                    manager,
                                    serviceStateManager
                                ).also {
                                    it.connect()
                                    it.startCommandPolling()
                                }
                            }
                        } else if (!ws.isConnected()) {
                            Log.w(TAG, "[HB-WD] WebSocket not connected, force reconnect")
                            ws.disconnect()
                            ws.connect()
                            ws.startCommandPolling()
                        } else {
                            Log.d(TAG, "[HB-WD] WebSocket connected, no reconnect needed")
                        }
                    } else {
                        heartbeatStaleStrike = 0
                        Log.d(TAG, "[HB-WD] heartbeat healthy, last success ${staleMs}ms ago")
                    }
                } else {
                    Log.d(TAG, "[HB-WD] waiting first heartbeat success")
                }
            }
        }
    }
}

package com.rvcs.android

import android.content.Context
import android.graphics.SurfaceTexture
import android.util.Log
import android.view.Surface
import kotlinx.coroutines.*

/**
 * 鏈嶅姟鐘舵€佺鐞嗗櫒
 * 缁熶竴绠＄悊鎽勫儚澶淬€佹帹娴佸拰缃戠粶杩炴帴鐘舵€?
 */
class ServiceStateManager(
    private val context: Context,
    private val registrar: DeviceRegistrar,  // 璁惧娉ㄥ唽鍣紝鐢ㄤ簬鑾峰彇LiveKit token
    private val liveKitServerUrl: String = "ws://YOUR_SERVER_IP:7880"
) {
    companion object {
        private const val TAG = "ServiceStateManager"
    }

    // 浠?DeviceRegistrar 鑾峰彇鏈嶅姟鍣║RL
    private val serverUrl: String
        get() = registrar.getServerUrl()

    // 鏍稿績缁勪欢
    private var streamManager: StreamManager? = null
    private var wakeLockHelper: WakeLockHelper? = null
    private var audioFocusHelper: AudioFocusHelper? = null

    // 璁惧淇℃伅
    private var deviceId: String? = null
    private var token: String? = null

    // 涓存椂瀛樺偍鐨?LiveKit 鏈嶅姟鍣║RL锛堜粠鏈嶅姟绔幏鍙栵級
    private var liveKitServerUrlFromToken: String = liveKitServerUrl

    // 鐘舵€?
    private val scope = CoroutineScope(Dispatchers.Default + Job())
    private var isCameraRunning = false
    private var isStreaming = false

    // 鐘舵€佸洖璋?
    var onCameraStateChange: ((Boolean, String?) -> Unit)? = null
    var onStreamingStateChange: ((Boolean, String?) -> Unit)? = null
    var onError: ((String) -> Unit)? = null
    
    /**
     * 鍒濆鍖栵紙甯﹁澶嘔D锛?
     */
    fun initialize(deviceId: String, token: String): Boolean {
        this.deviceId = deviceId
        this.token = token

        try {
            // 鍒濆鍖朩akeLock鍔╂墜
            wakeLockHelper = WakeLockHelper(context)

            // 鍒濆鍖栭煶棰戠劍鐐瑰姪鎵?
            audioFocusHelper = AudioFocusHelper(context)
            Log.d(TAG, "AudioFocusHelper initialized")

            // 鍒濆鍖栨祦绠＄悊鍣紙浣跨敤LiveKit锛岄粯璁?720p锛?
            // 浣跨敤 token 鎻愪緵鍥炶皟浠庢湇鍔＄鑾峰彇 LiveKit token
            val tokenProvider: suspend () -> String? = {
                // 浠庢湇鍔＄鑾峰彇 LiveKit token
                val tokenResult = registrar.getLiveKitToken()
                if (tokenResult == null) {
                    Log.e(TAG, "Failed to get LiveKit token")
                    null
                } else {
                    val (liveKitToken, _, serverUrl) = tokenResult

                    // 濡傛灉鏈嶅姟绔繑鍥炰簡涓嶅悓鐨?LiveKit 鏈嶅姟鍣║RL锛屾洿鏂版湰鍦伴厤缃?
                    if (serverUrl != liveKitServerUrl && liveKitServerUrlFromToken == liveKitServerUrl) {
                        liveKitServerUrlFromToken = serverUrl
                        Log.d(TAG, "LiveKit server URL updated from server: $serverUrl")
                    }

                    liveKitToken
                }
            }

            streamManager = StreamManager(
                context = context,
                liveKitServerUrl = liveKitServerUrl,
                deviceId = deviceId,
                tokenProvider = tokenProvider
            ).apply {
                setResolution(VideoResolution.RES_720P)
                onStreamingStart = {
                    isStreaming = true
                    onStreamingStateChange?.invoke(true, null)
                    Log.d(TAG, "LiveKit streaming started")
                }
                onStreamingStop = {
                    isStreaming = false
                    onStreamingStateChange?.invoke(false, null)
                    Log.d(TAG, "LiveKit streaming stopped")
                }
                onConnectionStateChange = { state ->
                    Log.d(TAG, "LiveKit connection state: $state")
                }
                onError = { error ->
                    Log.e(TAG, "StreamManager error: $error")
                    // 闃叉閫掑綊璋冪敤锛屽彧鍦ㄦ湭澶勭悊鏃惰Е鍙?
                    try {
                        this@ServiceStateManager.onError?.invoke(error)
                    } catch (e: Exception) {
                        Log.e(TAG, "Error in onError callback: ${e.message}")
                    }
                }
            }

            Log.i(TAG, "ServiceStateManager initialized with device ID")
            return true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize", e)
            return false
        }
    }
    
    /**
     * 鍒濆鍖栵紙涓嶅甫璁惧ID锛屼粎鐢ㄤ簬蹇冭烦鍜屽熀鏈姛鑳斤級
     */
    fun initializeWithoutCamera(token: String): Boolean {
        this.token = token
        
        try {
            // 鍒濆鍖朩akeLock鍔╂墜
            wakeLockHelper = WakeLockHelper(context)

            // 鍒濆鍖栭煶棰戠劍鐐瑰姪鎵?
            audioFocusHelper = AudioFocusHelper(context)
            Log.d(TAG, "AudioFocusHelper initialized")
            
            // 涓嶅垵濮嬪寲鎽勫儚澶寸鐞嗗櫒鍜屾祦绠＄悊鍣?
            // 杩欐牱鍙互鏀寔浠呮湁蹇冭烦鍔熻兘鐨勫満鏅?
            
            Log.i(TAG, "ServiceStateManager initialized without camera (heartbeat only)")
            return true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize without camera", e)
            return false
        }
    }
    
    /**
     * 鍚姩鎽勫儚澶达紙宸插簾寮冿細鎽勫儚澶寸敱 StreamManager 鍐呴儴绠＄悊锛?
     * @deprecated 浣跨敤 startStreaming() 浠ｆ浛
     */
    @Deprecated("Camera is managed by StreamManager")
    fun startCamera(): Boolean {
        Log.d(TAG, "=== startCamera called (deprecated, use startStreaming) ===")
        return true
    }
    
    /**
     * 鍋滄鎽勫儚澶达紙宸插簾寮冿細鎽勫儚澶寸敱 StreamManager 鍐呴儴绠＄悊锛?
     * @deprecated 浣跨敤 stopStreaming() 浠ｆ浛
     */
    @Deprecated("Camera is managed by StreamManager")
    fun stopCamera(): Boolean {
        Log.d(TAG, "=== stopCamera called (deprecated, use stopStreaming) ===")
        return stopStreaming()
    }
    
    /**
     * 寮€濮嬫帹娴?
     */
    fun startStreaming(): Boolean {
        Log.d(TAG, "=== startStreaming called ===")
        Log.d(TAG, "Current state - isStreaming: $isStreaming, streamManager: ${streamManager != null}")

        if (isStreaming) {
            Log.w(TAG, "Already streaming, skipping start")
            return true
        }

        return try {
            // 淇濇寔CPU鍞ら啋
            Log.d(TAG, "Acquiring wake lock...")
            wakeLockHelper?.startKeepAwake()

            // 璇锋眰闊抽鐒︾偣
            Log.d(TAG, "Requesting audio focus...")
            audioFocusHelper?.requestAudioFocus()
            val audioFocusResult = audioFocusHelper?.hasFocus() ?: false
            Log.d(TAG, "Audio focus result: $audioFocusResult")

            Log.d(TAG, "Starting streaming...")
            val result = streamManager?.startStreaming() ?: false
            Log.d(TAG, "Streaming start result: $result")

            if (result) {
                Log.i(TAG, "鉁?Streaming started successfully")
            } else {
                Log.e(TAG, "Failed to start streaming")
                wakeLockHelper?.stopKeepAwake()
                audioFocusHelper?.releaseAudioFocus()
            }

            result
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start streaming", e)
            wakeLockHelper?.stopKeepAwake()
            audioFocusHelper?.releaseAudioFocus()
            onError?.invoke("Streaming start error: ${e.message}")
            false
        }
    }
    
    /**
     * 鍋滄鎺ㄦ祦
     */
    fun stopStreaming(): Boolean {
        Log.d(TAG, "=== stopStreaming called ===")
        Log.d(TAG, "Current state - isStreaming: $isStreaming, streamManager: ${streamManager != null}")

        return try {
            Log.d(TAG, "Stopping streaming...")
            val result = streamManager?.stopStreaming() ?: true
            isStreaming = false
            Log.i(TAG, "Streaming stopped successfully: $result")

            wakeLockHelper?.stopKeepAwake()
            audioFocusHelper?.releaseAudioFocus()
            Log.d(TAG, "Released wake lock and audio focus")

            result
        } catch (e: Exception) {
            Log.e(TAG, "Failed to stop streaming", e)
            onError?.invoke("Streaming stop error: ${e.message}")
            false
        }
    }

    /**
     * 璁剧疆瑙嗛鍒嗚鲸鐜?
     */
    fun setResolution(resolution: VideoResolution): Boolean {
        return streamManager?.setResolution(resolution) ?: false
    }

    /**
     * 鑾峰彇褰撳墠鍒嗚鲸鐜?
     */
    fun getResolution(): VideoResolution? {
        return streamManager?.getResolution()
    }

    /**
     * 璁剧疆鐮佺巼
     */
    fun setBitrate(targetKbps: Int, minKbps: Int? = null, maxKbps: Int? = null): Boolean {
        return streamManager?.setBitrate(targetKbps, minKbps, maxKbps) ?: false
    }

    /**
     * 鍚敤/绂佺敤鑷€傚簲鐮佺巼
     */
    fun setAdaptiveBitrate(enabled: Boolean): Boolean {
        streamManager?.setAdaptiveBitrate(enabled)
        Log.d(TAG, "Adaptive bitrate ${if (enabled) "enabled" else "disabled"}")
        return true
    }

    /**
     * 閲嶅惎鎺ㄦ祦
     */
    fun restartCamera(): Boolean {
        return try {
            stopStreaming()
            Thread.sleep(500)
            startStreaming()
        } catch (e: Exception) {
            Log.e(TAG, "Failed to restart streaming", e)
            onError?.invoke("Streaming restart error: ${e.message}")
            false
        }
    }
    
    /**
     * 鍒濆鍖栨牳蹇冪粍浠?
     */
    fun initializeComponents() {
        Log.d(TAG, "Initializing core components...")

        // 鍒濆鍖栨帹娴佺鐞嗗櫒
        if (streamManager == null) {
            streamManager = StreamManager(
                context = context,
                liveKitServerUrl = liveKitServerUrl,
                deviceId = deviceId ?: "unknown",
                tokenProvider = { registrar.getLiveKitToken()?.first }
            )
            Log.d(TAG, "StreamManager initialized")
        }

        // 鍒濆鍖栧敜閱掗攣鍔╂墜
        if (wakeLockHelper == null) {
            wakeLockHelper = WakeLockHelper(context)
            Log.d(TAG, "WakeLockHelper initialized")
        }

        Log.i(TAG, "鉁?Core components initialization completed")
    }
    
    /**
     * 鑾峰彇娴佺鐞嗗櫒
     */
    fun getStreamManager(): StreamManager? {
        return streamManager
    }

    /**
     * 鑾峰彇鎽勫儚澶寸姸鎬?
     */
    fun getCameraState(): ServiceState {
        return ServiceState(
            isCameraRunning = isCameraRunning,
            isStreaming = isStreaming,
            streamType = null,
            isWakeLockHeld = wakeLockHelper?.isHolding() ?: false
        )
    }
    
    /**
     * 澶勭悊鎽勫儚澶撮敊璇?
     */
    fun handleCameraError() {
        Log.w(TAG, "Handling camera error by restarting streaming")
        // 閲嶅惎鎺ㄦ祦
        if (isStreaming) {
            stopStreaming()
            Thread.sleep(1000)
            startStreaming()
        }
    }
    
    /**
     * 娓呯悊璧勬簮
     */
    fun cleanup() {
        Log.d(TAG, "Cleaning up ServiceStateManager")

        // 鍋滄鎺ㄦ祦
        stopStreaming()

        // 娓呯悊 StreamManager
        streamManager?.dispose()
        streamManager = null

        wakeLockHelper?.cleanup()
        wakeLockHelper = null

        audioFocusHelper?.cleanup()
        audioFocusHelper = null

        scope.cancel()
    }
    
    /**
     * 鏈嶅姟鐘舵€佹暟鎹被
     */
    data class ServiceState(
        val isCameraRunning: Boolean,
        val isStreaming: Boolean,
        val streamType: String?,
        val isWakeLockHeld: Boolean
    )
}



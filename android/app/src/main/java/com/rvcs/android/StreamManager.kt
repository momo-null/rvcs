package com.rvcs.android

import android.content.Context
import android.util.Log
import java.util.concurrent.atomic.AtomicBoolean

/**
 * Stream manager based on AutoCameraStreamManager.
 */
class StreamManager(
    private val context: Context,
    private val liveKitServerUrl: String,
    private val deviceId: String,
    private val tokenProvider: suspend () -> String?,
    private var resolution: VideoResolution = VideoResolution.RES_720P
) {
    private val TAG = "StreamManager"

    private var autoCameraStreamManager: AutoCameraStreamManager? = null
    private val isStreaming = AtomicBoolean(false)

    var onStreamingStart: (() -> Unit)? = null
    var onStreamingStop: (() -> Unit)? = null
    var onConnectionStateChange: ((String) -> Unit)? = null
    var onError: ((String) -> Unit)? = null

    private var isCallingErrorCallback = false

    fun startStreaming(): Boolean {
        if (isStreaming.get()) {
            Log.w(TAG, "Already streaming")
            return false
        }

        Log.d(TAG, "Starting stream with resolution: ${resolution.displayName}")

        return try {
            autoCameraStreamManager = AutoCameraStreamManager(
                context = context,
                serverUrl = liveKitServerUrl,
                tokenProvider = tokenProvider
            ).apply {
                onConnectionStateChange = { state ->
                    Log.d(TAG, "LiveKit state: $state")
                    try {
                        this@StreamManager.onConnectionStateChange?.invoke(state)
                    } catch (e: Exception) {
                        Log.e(TAG, "onConnectionStateChange callback error: ${e.message}", e)
                    }

                    when (state) {
                        "CONNECTED", "PUBLISHING" -> {
                            if (!isStreaming.get()) {
                                isStreaming.set(true)
                                try {
                                    this@StreamManager.onStreamingStart?.invoke()
                                } catch (e: Exception) {
                                    Log.e(TAG, "onStreamingStart callback error: ${e.message}", e)
                                    invokeErrorSafely("onStreamingStart callback error: ${e.message}")
                                }
                            }
                        }

                        "DISCONNECTED", "FAILED" -> {
                            if (isStreaming.getAndSet(false)) {
                                safeInvokeStreamingStop()
                            }
                        }
                    }
                }

                onError = { error ->
                    Log.e(TAG, "LiveKit error: $error")
                    invokeErrorSafely("LiveKit error: $error")
                    if (isStreaming.getAndSet(false)) {
                        safeInvokeStreamingStop()
                    }
                }
            }

            autoCameraStreamManager!!.startStreaming()
            Log.i(TAG, "Stream start requested")
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start streaming: ${e.message}", e)
            invokeErrorSafely("stream start failed: ${e.message}")
            false
        }
    }

    fun stopStreaming(): Boolean {
        Log.d(TAG, "Stopping stream...")

        return try {
            val result = autoCameraStreamManager?.stopStreaming() ?: true
            isStreaming.set(false)
            autoCameraStreamManager = null
            safeInvokeStreamingStop()
            result
        } catch (e: Exception) {
            Log.e(TAG, "Failed to stop streaming: ${e.message}", e)
            isStreaming.set(false)
            autoCameraStreamManager = null
            invokeErrorSafely("stream stop failed: ${e.message}")
            false
        }
    }

    fun isStreaming(): Boolean = isStreaming.get()

    fun setResolution(newResolution: VideoResolution): Boolean {
        return try {
            if (isStreaming.get()) {
                Log.w(TAG, "Cannot change resolution while streaming")
                return false
            }
            resolution = newResolution
            Log.d(TAG, "Resolution set to: ${newResolution.displayName}")
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to set resolution: ${e.message}", e)
            false
        }
    }

    fun getResolution(): VideoResolution = resolution

    fun changeResolution(newResolution: VideoResolution): Boolean {
        return try {
            if (newResolution == resolution) return true
            Log.d(TAG, "Changing resolution to: ${newResolution.displayName}")
            val wasStreaming = isStreaming.get()
            resolution = newResolution
            if (wasStreaming) {
                stopStreaming()
                startStreaming()
            }
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to change resolution: ${e.message}", e)
            false
        }
    }

    fun setAdaptiveBitrate(enabled: Boolean) {
        Log.d(TAG, "Adaptive bitrate ${if (enabled) "enabled" else "disabled"}")
    }

    @Suppress("UNUSED_PARAMETER")
    fun setBitrate(targetKbps: Int, minKbps: Int? = null, maxKbps: Int? = null): Boolean {
        Log.d(TAG, "Bitrate set: target=$targetKbps kbps (disabled in current version)")
        return true
    }

    fun dispose() {
        Log.d(TAG, "Disposing StreamManager")
        stopStreaming()
        autoCameraStreamManager?.dispose()
        autoCameraStreamManager = null
    }

    private fun invokeErrorSafely(message: String) {
        if (!isCallingErrorCallback) {
            isCallingErrorCallback = true
            try {
                onError?.invoke(message)
            } catch (e: Exception) {
                Log.e(TAG, "Exception in onError callback: ${e.message}", e)
            } finally {
                isCallingErrorCallback = false
            }
        }
    }

    private fun safeInvokeStreamingStop() {
        try {
            onStreamingStop?.invoke()
        } catch (e: Exception) {
            Log.e(TAG, "Exception in onStreamingStop callback: ${e.message}", e)
            invokeErrorSafely("onStreamingStop callback error: ${e.message}")
        }
    }
}

package com.rvcs.android

import android.content.Context
import android.media.AudioManager
import android.util.Log
import io.livekit.android.LiveKit
import io.livekit.android.events.RoomEvent
import io.livekit.android.events.collect
import io.livekit.android.room.Room
import io.livekit.android.room.track.CameraPosition
import io.livekit.android.room.track.LocalVideoTrack
import io.livekit.android.room.track.LocalVideoTrackOptions
import io.livekit.android.room.track.VideoPreset169
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext

class AutoCameraStreamManager(
    private val context: Context,
    private val serverUrl: String,
    private val tokenProvider: suspend () -> String?
) {

    companion object {
        private const val TAG = "AutoCameraStreamManager"
    }

    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private val stopMutex = Mutex()

    private var room: Room? = null
    private var eventsJob: Job? = null
    private var reconnectJob: Job? = null
    private var publishedVideoTrack: LocalVideoTrack? = null
    private var shouldReconnect = false
    private val audioManager: AudioManager by lazy {
        context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
    }
    private var originalAudioMode: Int = AudioManager.MODE_NORMAL
    private var originalSpeakerphoneOn: Boolean = false
    private var originalBluetoothScoOn: Boolean = false
    private var audioRouteForced: Boolean = false

    var onConnected: (() -> Unit)? = null
    var onDisconnected: (() -> Unit)? = null
    var onConnectionStateChange: ((String) -> Unit)? = null
    var onError: ((String) -> Unit)? = null

    fun startStreaming() {
        shouldReconnect = true
        scope.launch {
            connectInternal()
        }
    }

    private suspend fun connectInternal() {
        try {
            if (room != null) {
                Log.w(TAG, "Existing room detected, shutting down before reconnect")
                stopStreamingInternal()
            }

            val token = tokenProvider()
            if (token.isNullOrEmpty()) {
                onError?.invoke("token is empty")
                return
            }

            room = LiveKit.create(context)
            observeRoomEvents(room!!)

            Log.i(TAG, "Connecting to LiveKit: $serverUrl")
            room?.connect(serverUrl, token)
        } catch (e: Exception) {
            Log.e(TAG, "Connect error", e)
            onError?.invoke(e.message ?: "connect error")
            if (shouldReconnect) {
                scheduleReconnect()
            }
        }
    }

    private fun observeRoomEvents(targetRoom: Room) {
        eventsJob = scope.launch {
            targetRoom.events.collect { event ->
                when (event) {
                    is RoomEvent.Connected -> {
                        Log.i(TAG, "LiveKit connected")
                        onConnectionStateChange?.invoke("CONNECTED")
                        onConnected?.invoke()

                        try {
                            forceSpeakerRoute()
                            // Keep audio path simple: enable mic directly.
                            targetRoom.localParticipant.setMicrophoneEnabled(true)

                            // Create and publish one explicit camera track; track handle is saved
                            // so stop flow can always release the camera synchronously.
                            val options = LocalVideoTrackOptions(
                                isScreencast = false,
                                deviceId = null,
                                position = CameraPosition.BACK,
                                captureParams = VideoPreset169.H720.capture
                            )

                            val videoTrack = targetRoom.localParticipant.createVideoTrack(
                                name = "CameraBack",
                                options = options
                            )
                            targetRoom.localParticipant.publishVideoTrack(videoTrack)
                            publishedVideoTrack = videoTrack
                            // For LiveKit 2.10.x, explicit camera enable is still required
                            // in this create+publish flow to actually start capture frames.
                            targetRoom.localParticipant.setCameraEnabled(true)

                            Log.i(TAG, "Camera track published")
                        } catch (e: Exception) {
                            Log.e(TAG, "Failed to publish local tracks", e)
                            onError?.invoke("publish track failed: ${e.message}")
                        }
                    }

                    is RoomEvent.Disconnected -> {
                        Log.w(TAG, "LiveKit disconnected")
                        onConnectionStateChange?.invoke("DISCONNECTED")
                        onDisconnected?.invoke()

                        if (shouldReconnect) {
                            scheduleReconnect()
                        }
                    }

                    else -> Unit
                }
            }
        }
    }

    private fun scheduleReconnect() {
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            delay(3000)
            if (shouldReconnect) {
                Log.w(TAG, "Reconnecting LiveKit...")
                connectInternal()
            }
        }
    }

    private suspend fun stopStreamingInternal(): Boolean {
        return stopStreamingSuspend("internal")
    }

    private suspend fun stopStreamingSuspend(source: String): Boolean {
        return stopMutex.withLock {
            Log.d(TAG, "Stopping streaming ($source)")

            shouldReconnect = false
            reconnectJob?.cancel()
            reconnectJob = null

            val currentEventsJob = eventsJob
            eventsJob = null
            currentEventsJob?.cancelAndJoin()

            val currentRoom = room
            if (currentRoom == null) {
                publishedVideoTrack = null
                onConnectionStateChange?.invoke("DISCONNECTED")
                return@withLock true
            }

            try {
                withContext(Dispatchers.Main.immediate) {
                    try {
                        publishedVideoTrack?.let { track ->
                            try {
                                currentRoom.localParticipant.unpublishTrack(track)
                                Log.d(TAG, "Local video track unpublished")
                            } catch (e: Exception) {
                                Log.w(TAG, "Unpublish video track warning: ${e.message}")
                            }
                            track.stop()
                            Log.d(TAG, "Local video track stopped")
                        }
                    } catch (e: Exception) {
                        Log.w(TAG, "Local video track stop warning: ${e.message}")
                    } finally {
                        publishedVideoTrack = null
                    }

                    try {
                        currentRoom.localParticipant.setCameraEnabled(false)
                    } catch (e: Exception) {
                        Log.w(TAG, "setCameraEnabled(false) warning: ${e.message}")
                    }

                    try {
                        currentRoom.localParticipant.setMicrophoneEnabled(false)
                    } catch (e: Exception) {
                        Log.w(TAG, "setMicrophoneEnabled(false) warning: ${e.message}")
                    }
                }

                delay(300)

                withContext(Dispatchers.Main.immediate) {
                    currentRoom.disconnect()
                }

                delay(300)
                onConnectionStateChange?.invoke("DISCONNECTED")
                Log.i(TAG, "Streaming resources released")
                true
            } catch (e: Exception) {
                Log.e(TAG, "Failed to release streaming resources", e)
                onError?.invoke("stop streaming failed: ${e.message}")
                false
            } finally {
                restoreAudioRoute()
                room = null
                publishedVideoTrack = null
            }
        }
    }

    fun stopStreaming(): Boolean {
        return runBlocking {
            stopStreamingSuspend("public")
        }
    }

    fun enableCamera(enable: Boolean) {
        scope.launch {
            room?.localParticipant?.setCameraEnabled(enable)
        }
    }

    fun enableMicrophone(enable: Boolean) {
        scope.launch {
            room?.localParticipant?.setMicrophoneEnabled(enable)
        }
    }

    fun dispose() {
        Log.d(TAG, "Disposing AutoCameraStreamManager")
        runBlocking {
            stopStreamingSuspend("dispose")
        }

        scope.cancel()

        onConnected = null
        onDisconnected = null
        onConnectionStateChange = null
        onError = null
    }

    // Force voice playback to loud speaker during talkback/intercom.
    private fun forceSpeakerRoute() {
        if (audioRouteForced) return
        try {
            originalAudioMode = audioManager.mode
            originalSpeakerphoneOn = audioManager.isSpeakerphoneOn
            originalBluetoothScoOn = audioManager.isBluetoothScoOn

            audioManager.mode = AudioManager.MODE_IN_COMMUNICATION
            audioManager.stopBluetoothSco()
            audioManager.isBluetoothScoOn = false
            audioManager.isSpeakerphoneOn = true

            audioRouteForced = true
            Log.i(TAG, "Audio routed to speaker")
        } catch (e: Exception) {
            Log.w(TAG, "Failed to force speaker route: ${e.message}")
        }
    }

    // Restore previous audio route to avoid leaking routing state.
    private fun restoreAudioRoute() {
        if (!audioRouteForced) return
        try {
            audioManager.isSpeakerphoneOn = originalSpeakerphoneOn
            audioManager.isBluetoothScoOn = originalBluetoothScoOn
            audioManager.mode = originalAudioMode
            audioRouteForced = false
            Log.i(TAG, "Audio route restored")
        } catch (e: Exception) {
            Log.w(TAG, "Failed to restore audio route: ${e.message}")
        }
    }
}

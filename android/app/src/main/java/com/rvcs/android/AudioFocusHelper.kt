package com.rvcs.android

import android.content.Context
import android.media.AudioAttributes
import android.media.AudioFocusRequest
import android.media.AudioManager
import android.os.Build
import android.util.Log

/**
 * 音频焦点辅助类
 * 用于管理音频焦点，确保熄屏后音频能正常播放
 */
class AudioFocusHelper(private val context: Context) {
    private val TAG = "AudioFocusHelper"

    private val audioManager: AudioManager by lazy {
        context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
    }

    private var audioFocusRequest: AudioFocusRequest? = null
    private var hasAudioFocus = false

    /**
     * 请求音频焦点
     */
    fun requestAudioFocus(): Boolean {
        try {
            Log.d(TAG, "Requesting audio focus...")

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                // Android 8.0+ 使用 AudioFocusRequest
                val audioAttributes = AudioAttributes.Builder()
                    .setUsage(AudioAttributes.USAGE_MEDIA)
                    .setContentType(AudioAttributes.CONTENT_TYPE_SPEECH)
                    .build()

                audioFocusRequest = AudioFocusRequest.Builder(AudioManager.AUDIOFOCUS_GAIN)
                    .setAudioAttributes(audioAttributes)
                    .setAcceptsDelayedFocusGain(true)
                    .setOnAudioFocusChangeListener(audioFocusChangeListener)
                    .build()

                val result = audioManager.requestAudioFocus(audioFocusRequest!!)
                hasAudioFocus = result == AudioManager.AUDIOFOCUS_REQUEST_GRANTED

                Log.d(TAG, "Audio focus request result: $result, hasFocus: $hasAudioFocus")
                return hasAudioFocus
            } else {
                // Android 8.0 以下使用旧 API
                @Suppress("DEPRECATION")
                val result = audioManager.requestAudioFocus(
                    audioFocusChangeListener,
                    AudioManager.STREAM_MUSIC,
                    AudioManager.AUDIOFOCUS_GAIN
                )
                hasAudioFocus = result == AudioManager.AUDIOFOCUS_REQUEST_GRANTED

                Log.d(TAG, "Audio focus request result (old API): $result, hasFocus: $hasAudioFocus")
                return hasAudioFocus
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to request audio focus", e)
            return false
        }
    }

    /**
     * 释放音频焦点
     */
    fun releaseAudioFocus() {
        try {
            Log.d(TAG, "Releasing audio focus...")

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                audioFocusRequest?.let {
                    audioManager.abandonAudioFocusRequest(it)
                    audioFocusRequest = null
                }
            } else {
                @Suppress("DEPRECATION")
                audioManager.abandonAudioFocus(audioFocusChangeListener)
            }

            hasAudioFocus = false
            Log.d(TAG, "Audio focus released")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to release audio focus", e)
        }
    }

    /**
     * 检查是否持有音频焦点
     */
    fun hasFocus(): Boolean = hasAudioFocus

    /**
     * 音频焦点变化监听器
     */
    private val audioFocusChangeListener = AudioManager.OnAudioFocusChangeListener { focusChange ->
        when (focusChange) {
            AudioManager.AUDIOFOCUS_GAIN -> {
                Log.i(TAG, "Audio focus GAINED")
                hasAudioFocus = true
                // 可以恢复或开始播放
            }
            AudioManager.AUDIOFOCUS_LOSS -> {
                Log.w(TAG, "Audio focus LOST (permanent)")
                hasAudioFocus = false
                // 永久丢失，应该停止播放
            }
            AudioManager.AUDIOFOCUS_LOSS_TRANSIENT -> {
                Log.w(TAG, "Audio focus LOST (transient)")
                hasAudioFocus = false
                // 临时丢失，应该暂停播放
            }
            AudioManager.AUDIOFOCUS_LOSS_TRANSIENT_CAN_DUCK -> {
                Log.w(TAG, "Audio focus LOST (transient can duck)")
                hasAudioFocus = false
                // 临时丢失，可以降低音量
            }
        }
    }

    /**
     * 获取当前音量
     */
    fun getMusicVolume(): Int {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            audioManager.getStreamVolume(AudioManager.STREAM_MUSIC)
        } else {
            @Suppress("DEPRECATION")
            audioManager.getStreamVolume(AudioManager.STREAM_MUSIC)
        }
    }

    /**
     * 获取最大音量
     */
    fun getMaxMusicVolume(): Int {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            audioManager.getStreamMaxVolume(AudioManager.STREAM_MUSIC)
        } else {
            @Suppress("DEPRECATION")
            audioManager.getStreamMaxVolume(AudioManager.STREAM_MUSIC)
        }
    }

    /**
     * 检查音频是否静音
     */
    fun isMusicMuted(): Boolean {
        return getMusicVolume() == 0
    }

    /**
     * 清理资源
     */
    fun cleanup() {
        releaseAudioFocus()
        Log.d(TAG, "AudioFocusHelper cleaned up")
    }
}

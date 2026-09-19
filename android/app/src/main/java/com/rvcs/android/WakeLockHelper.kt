package com.rvcs.android

import android.content.Context
import android.os.PowerManager
import android.util.Log

/**
 * Helper for a process-level partial wakelock.
 * Keeps CPU alive when screen is off; must be explicitly released.
 */
class WakeLockHelper(private val context: Context) {

    companion object {
        private const val TAG = "WakeLockHelper"
        private const val WAKE_LOCK_TAG = "rvcs:cameraservice"
    }

    private val powerManager: PowerManager by lazy {
        context.getSystemService(Context.POWER_SERVICE) as PowerManager
    }

    private var partialWakeLock: PowerManager.WakeLock? = null
    private var isWakeLockHeld = false

    private fun acquirePartialWakeLock(): Boolean {
        return try {
            if (partialWakeLock == null) {
                partialWakeLock = powerManager.newWakeLock(
                    PowerManager.PARTIAL_WAKE_LOCK,
                    WAKE_LOCK_TAG
                ).apply {
                    setReferenceCounted(false)
                }
            }

            val lock = partialWakeLock
            if (!isWakeLockHeld || lock?.isHeld != true) {
                // No timeout: release is controlled by service lifecycle.
                lock?.acquire()
                isWakeLockHeld = true
                Log.d(TAG, "WakeLock acquired")
            } else {
                Log.d(TAG, "WakeLock already held")
            }
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to acquire WakeLock", e)
            false
        }
    }

    private fun releaseWakeLock() {
        try {
            val lock = partialWakeLock
            if (isWakeLockHeld && lock?.isHeld == true) {
                lock.release()
                isWakeLockHeld = false
                Log.d(TAG, "WakeLock released")
            } else {
                isWakeLockHeld = false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to release WakeLock", e)
            isWakeLockHeld = false
        }
    }

    fun startKeepAwake(): Boolean = acquirePartialWakeLock()

    fun stopKeepAwake() {
        releaseWakeLock()
    }

    fun isHolding(): Boolean {
        return isWakeLockHeld && partialWakeLock?.isHeld == true
    }

    fun cleanup() {
        releaseWakeLock()
        partialWakeLock = null
        Log.d(TAG, "WakeLockHelper cleaned up")
    }
}

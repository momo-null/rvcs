package com.rvcs.android

import android.content.Context
import android.net.wifi.WifiManager
import android.util.Log

/**
 * Keep Wi-Fi active during screen-off for heartbeat/websocket traffic.
 */
class WifiLockHelper(context: Context) {

    companion object {
        private const val TAG = "WifiLockHelper"
        private const val WIFI_LOCK_TAG = "rvcs:wifi-lock"
    }

    private val wifiManager = context.applicationContext.getSystemService(Context.WIFI_SERVICE) as WifiManager
    private var wifiLock: WifiManager.WifiLock? = null
    private var isHeld = false

    fun acquire(): Boolean {
        return try {
            if (wifiLock == null) {
                wifiLock = wifiManager.createWifiLock(WifiManager.WIFI_MODE_FULL_HIGH_PERF, WIFI_LOCK_TAG).apply {
                    setReferenceCounted(false)
                }
            }
            val lock = wifiLock
            if (!isHeld || lock?.isHeld != true) {
                lock?.acquire()
                isHeld = true
                Log.d(TAG, "WifiLock acquired")
            }
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to acquire WifiLock", e)
            false
        }
    }

    fun release() {
        try {
            val lock = wifiLock
            if (isHeld && lock?.isHeld == true) {
                lock.release()
                Log.d(TAG, "WifiLock released")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to release WifiLock", e)
        } finally {
            isHeld = false
        }
    }

    fun cleanup() {
        release()
        wifiLock = null
    }
}


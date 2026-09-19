package com.rvcs.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log

/**
 * 开机广播接收器
 * 设备启动后自动启动相机服务
 */
class BootReceiver : BroadcastReceiver() {
    
    companion object {
        private const val TAG = "BootReceiver"
    }
    
    override fun onReceive(context: Context, intent: Intent?) {
        if (intent?.action == Intent.ACTION_BOOT_COMPLETED) {
            Log.d(TAG, "Boot completed received, starting service")
            
            try {
                // 启动前台相机服务
                val serviceIntent = Intent(context, ForegroundCameraService::class.java).apply {
                    action = ForegroundCameraService.ACTION_START
                }
                
                if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
                    context.startForegroundService(serviceIntent)
                } else {
                    context.startService(serviceIntent)
                }
                
                Log.i(TAG, "Service started from boot receiver")
            } catch (e: Exception) {
                Log.e(TAG, "Failed to start service from boot receiver", e)
            }
        }
    }
}

package com.rvcs.android

import android.Manifest
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.os.PowerManager
import android.provider.Settings
import android.util.Log
import androidx.core.content.ContextCompat

/**
 * 电池优化和Doze模式辅助类
 * 用于确保应用在后台持续运行
 */
class BatteryOptimizationHelper(private val context: Context) {
    
    companion object {
        private const val TAG = "BatteryOptHelper"
        const val REQUEST_IGNORE_BATTERY_OPTIMIZATIONS = 1001
    }
    
    private val powerManager: PowerManager by lazy {
        context.getSystemService(Context.POWER_SERVICE) as PowerManager
    }
    
    /**
     * 检查是否忽略电池优化
     */
    fun isIgnoringBatteryOptimizations(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            powerManager.isIgnoringBatteryOptimizations(context.packageName)
        } else {
            true // Android 6.0以下不需要处理
        }
    }
    
    /**
     * 请求忽略电池优化
     * 需要在Activity中调用此方法并处理结果
     */
    fun requestIgnoreBatteryOptimizations(): Intent? {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            if (!isIgnoringBatteryOptimizations()) {
                Intent().apply {
                    action = Settings.ACTION_REQUEST_IGNORE_BATTERY_OPTIMIZATIONS
                    data = Uri.parse("package:${context.packageName}")
                }
            } else {
                null
            }
        } else {
            null
        }
    }
    
    /**
     * 打开电池优化设置页面
     */
    fun openBatteryOptimizationSettings(): Intent {
        return Intent().apply {
            action = Settings.ACTION_IGNORE_BATTERY_OPTIMIZATION_SETTINGS
        }
    }
    
    /**
     * 检查是否在Doze模式白名单中
     */
    fun isInDozeWhitelist(): Boolean {
        return isIgnoringBatteryOptimizations()
    }
    
    /**
     * 检查应用是否被限制后台运行
     */
    fun isBackgroundRestricted(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            val activityManager = context.getSystemService(Context.ACTIVITY_SERVICE) as android.app.ActivityManager
            activityManager.isBackgroundRestricted
        } else {
            false
        }
    }
    
    /**
     * 打开应用设置页面
     * 用于手动设置后台运行权限
     */
    fun openAppSettings(): Intent {
        return Intent().apply {
            action = Settings.ACTION_APPLICATION_DETAILS_SETTINGS
            data = Uri.parse("package:${context.packageName}")
        }
    }
    
    /**
     * 检查是否有REQUEST_IGNORE_BATTERY_OPTIMIZATIONS权限
     */
    fun hasIgnoreBatteryOptimizationsPermission(): Boolean {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            ContextCompat.checkSelfPermission(
                context,
                Manifest.permission.REQUEST_IGNORE_BATTERY_OPTIMIZATIONS
            ) == PackageManager.PERMISSION_GRANTED
        } else {
            true
        }
    }
    
    /**
     * 获取优化建议
     */
    fun getOptimizationTips(): List<String> {
        val tips = mutableListOf<String>()
        
        if (!isIgnoringBatteryOptimizations()) {
            tips.add("建议：将应用添加到电池优化白名单")
        }
        
        if (isBackgroundRestricted()) {
            tips.add("建议：关闭应用的后台运行限制")
        }
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            tips.add("建议：在系统设置中允许应用在后台运行")
        }
        
        return tips
    }
    
    /**
     * 打印当前优化状态
     */
    fun logOptimizationStatus() {
        Log.d(TAG, "=== 电池优化状态 ===")
        Log.d(TAG, "忽略电池优化: ${isIgnoringBatteryOptimizations()}")
        Log.d(TAG, "Doze白名单: ${isInDozeWhitelist()}")
        Log.d(TAG, "后台限制: ${isBackgroundRestricted()}")
        Log.d(TAG, "权限检查: ${hasIgnoreBatteryOptimizationsPermission()}")
        Log.d(TAG, "===================")
    }
}

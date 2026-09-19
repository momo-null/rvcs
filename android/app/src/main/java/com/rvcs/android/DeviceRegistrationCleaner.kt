package com.rvcs.android

import android.content.Context
import android.content.SharedPreferences
import android.util.Log

/**
 * 设备注册状态清理工具
 * 用于调试时清除不完整的注册状态
 */
class DeviceRegistrationCleaner(private val context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
    
    companion object {
        private const val TAG = "RegistrationCleaner"
    }
    
    /**
     * 检查当前注册状态的完整性
     */
    fun checkRegistrationStatus(): Map<String, Any?> {
        val deviceId = prefs.getString("device_id", null)
        val deviceToken = prefs.getString("device_token", null)
        val refreshToken = prefs.getString("refresh_token", null)
        val isRegistered = prefs.getBoolean("is_registered", false)
        val serverUrl = prefs.getString("server_url", null)
        
        val status = mapOf(
            "deviceId" to deviceId,
            "deviceToken" to deviceToken,
            "refreshToken" to refreshToken,
            "isRegistered" to isRegistered,
            "serverUrl" to serverUrl,
            "isComplete" to (deviceId != null && deviceToken != null)
        )
        
        Log.d(TAG, "Registration status check:")
        Log.d(TAG, "- deviceId: ${deviceId ?: "null"}")
        Log.d(TAG, "- deviceToken: ${deviceToken?.take(20)}...")
        Log.d(TAG, "- refreshToken: ${refreshToken?.take(20)}...")
        Log.d(TAG, "- isRegistered: $isRegistered")
        Log.d(TAG, "- serverUrl: $serverUrl")
        Log.d(TAG, "- isComplete: ${status["isComplete"]}")
        
        return status
    }
    
    /**
     * 清除所有注册相关信息
     */
    fun clearRegistration() {
        Log.d(TAG, "Clearing all registration data...")
        
        prefs.edit()
            .remove("device_id")
            .remove("device_token")
            .remove("refresh_token")
            .remove("is_registered")
            .remove("server_url")
            .remove("device_name")
            .apply()
            
        Log.d(TAG, "Registration data cleared successfully")
    }
    
    /**
     * 强制设置完整注册状态（仅用于测试）
     */
    fun forceSetRegistration(deviceId: String, deviceToken: String, refreshToken: String = "") {
        Log.d(TAG, "Force setting registration data...")
        Log.d(TAG, "Device ID: $deviceId")
        Log.d(TAG, "Device Token: ${deviceToken.take(20)}...")
        
        prefs.edit()
            .putString("device_id", deviceId)
            .putString("device_token", deviceToken)
            .putString("refresh_token", refreshToken)
            .putBoolean("is_registered", true)
            .apply()
            
        Log.d(TAG, "Registration data set successfully")
    }
}
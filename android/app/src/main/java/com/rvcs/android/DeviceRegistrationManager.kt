package com.rvcs.android

import android.util.Log
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import org.json.JSONObject
import java.io.IOException

/**
 * 设备注册管理器
 * 支持注册码验证的设备注册流程
 */
class DeviceRegistrationManager {
    private val TAG = "DeviceRegistration"
    private val client = OkHttpClient()
    private val baseUrl = "https://your-server-address:28443/api/v1/device" // Android端设备接口
    
    /**
     * 使用注册码注册设备
     */
    fun registerWithCode(
        registrationCode: String,
        deviceInfo: DeviceInfo,
        callback: (Boolean, String?) -> Unit
    ) {
        try {
            // 构建注册请求
            val requestBody = JSONObject().apply {
                put("registration_code", registrationCode)
                put("device_info", JSONObject().apply {
                    put("name", deviceInfo.name)
                    put("type", deviceInfo.type)
                    put("model", deviceInfo.model)
                    put("location", deviceInfo.location)
                    put("description", deviceInfo.description)
                    put("os_version", deviceInfo.osVersion)
                    put("app_version", deviceInfo.appVersion)
                })
            }
            
            val request = Request.Builder()
                .url("$baseUrl/devices/register")
                .post(RequestBody.create(
                    "application/json".toMediaType(), 
                    requestBody.toString()
                ))
                .build()
            
            client.newCall(request).enqueue(object : Callback {
                override fun onFailure(call: Call, e: IOException) {
                    Log.e(TAG, "Registration failed", e)
                    callback(false, e.message)
                }
                
                override fun onResponse(call: Call, response: Response) {
                    try {
                        val responseBody = response.body?.string()
                        if (response.isSuccessful && responseBody != null) {
                            val jsonResponse = JSONObject(responseBody)
                            if (jsonResponse.has("data")) {
                                val deviceData = jsonResponse.getJSONObject("data")
                                val deviceId = deviceData.getString("id")
                                Log.d(TAG, "Device registered successfully: $deviceId")
                                callback(true, deviceId)
                            } else {
                                callback(false, "Invalid response format")
                            }
                        } else {
                            val errorMessage = responseBody ?: "Registration failed"
                            Log.e(TAG, "Registration error: $errorMessage")
                            callback(false, errorMessage)
                        }
                    } catch (e: Exception) {
                        Log.e(TAG, "Response parsing failed", e)
                        callback(false, e.message)
                    }
                }
            })
        } catch (e: Exception) {
            Log.e(TAG, "Registration request preparation failed", e)
            callback(false, e.message)
        }
    }
    
    /**
     * 设备信息数据类
     */
    data class DeviceInfo(
        val name: String,
        val type: String,
        val model: String? = null,
        val location: String? = null,
        val description: String? = null,
        val osVersion: String? = null,
        val appVersion: String? = null
    )
    
    /**
     * 验证注册码格式
     */
    fun isValidRegistrationCode(code: String): Boolean {
        // 基本格式验证
        if (code.length < 58) return false
        if (!code.startsWith("RVCS-")) return false
        
        // 更详细的验证可以在这里实现
        return true
    }
    
    /**
     * 解析注册码信息
     */
    fun parseRegistrationCode(code: String): RegistrationCodeInfo? {
        if (!isValidRegistrationCode(code)) return null
        
        return try {
            val parts = code.split("-")
            if (parts.size >= 4) {
                RegistrationCodeInfo(
                    prefix = parts[0],
                    timestamp = parts[1],
                    randomPart = parts[2],
                    checksum = parts[3]
                )
            } else {
                null
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to parse registration code", e)
            null
        }
    }
    
    /**
     * 注册码信息数据类
     */
    data class RegistrationCodeInfo(
        val prefix: String,
        val timestamp: String,
        val randomPart: String,
        val checksum: String
    )
}
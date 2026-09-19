package com.rvcs.android

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import java.security.KeyPairGenerator
import java.util.Base64

class DeviceRegistrar(private val context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
    private val client = SSLUtils.createUnsafeOkHttpClient()

    companion object {
        private const val TAG = "DeviceRegistrar"
    }

    fun isRegistered(): Boolean {
        // 检查设备是否已注册
        val deviceId = prefs.getString("device_id", null)
        val deviceToken = prefs.getString("device_token", null)
        val isRegisteredFlag = prefs.getBoolean("is_registered", false)
        
        // 判断逻辑：
        // 1. 如果有deviceId且有有效的device_token，则认为已注册
        // 2. 或者如果isRegisteredFlag为true，则认为已注册
        val hasValidDeviceId = deviceId != null
        val hasValidToken = deviceToken != null
        val isMarkedAsRegistered = isRegisteredFlag
        
        val isRegistered = (hasValidDeviceId && hasValidToken) || isMarkedAsRegistered
        
        Log.d(TAG, "isRegistered check:")
        Log.d(TAG, "- deviceId: $deviceId (exists: $hasValidDeviceId)")
        Log.d(TAG, "- deviceToken: ${if (deviceToken != null) deviceToken.take(20) + "..." else "null"} (exists: $hasValidToken)")
        Log.d(TAG, "- isRegisteredFlag: $isRegisteredFlag")
        Log.d(TAG, "- Final result: $isRegistered")
        
        return isRegistered
    }

    fun getDeviceId(): String? = prefs.getString("device_id", null)

    fun setDeviceId(deviceId: String) {
        prefs.edit().putString("device_id", deviceId).apply()
        Log.d(TAG, "Device ID updated to: $deviceId")
    }

    fun getAccessToken(): String? {
        // 使用device_token
        val deviceToken = prefs.getString("device_token", null)
        Log.d(TAG, "getAccessToken: ${deviceToken?.take(20)}...")
        return deviceToken
    }

    fun getRefreshToken(): String? = prefs.getString("refresh_token", null)

    fun saveTokens(deviceId: String, accessToken: String, refreshToken: String) {
        prefs.edit()
            .putString("device_id", deviceId)
            .putString("device_token", accessToken)  // 统一使用device_token作为存储键名
            .putString("refresh_token", refreshToken)
            .putBoolean("is_registered", true)  // 确保设置注册状态
            .apply()
    }

    fun clearTokens() {
        prefs.edit()
            .remove("device_id")
            .remove("device_token")
            .remove("refresh_token")
            .remove("is_registered")
            .apply()
        Log.d(TAG, "Tokens cleared")
    }

    fun getServerUrl(): String {
        // 优先从 SharedPreferences 读取用户自定义的服务器地址
        return prefs.getString("server_url", null)
            ?: context.getString(R.string.server_base_url)
    }

    suspend fun refreshToken(): Boolean {
        val refreshToken = getRefreshToken() ?: return false

        try {
            val server = getServerUrl()
            val url = "$server/api/v1/device/refresh"

            val bodyJson = JSONObject()
            bodyJson.put("refresh_token", refreshToken)

            val body = bodyJson.toString().toRequestBody("application/json".toMediaTypeOrNull())
            val req = Request.Builder().url(url).post(body).build()
            client.newCall(req).execute().use { resp ->
                if (resp.isSuccessful) {
                    val text = resp.body?.string() ?: ""
                    Log.d(TAG, "Token refresh response: $text")
                    val json = JSONObject(text)
                    val data = json.getJSONObject("data")
                    val access = data.getString("device_token")
                    val refresh = data.getString("refresh_token")

                    val deviceId = getDeviceId()!!
                    saveTokens(deviceId, access, refresh)
                    Log.d(TAG, "Token refreshed successfully")
                    return true
                } else {
                    Log.e(TAG, "Token refresh failed: ${resp.code}")
                    clearTokens()
                    return false
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Token refresh error", e)
            clearTokens()
            return false
        }
    }

    /**
     * 获取 LiveKit Token
     * 从服务端获取用于连接 LiveKit 服务器的 token
     * @return Triple(token, roomId, serverUrl) 或 null
     */
    suspend fun getLiveKitToken(): Triple<String, String, String>? {
        val deviceId = getDeviceId() ?: run {
            Log.e(TAG, "Cannot get LiveKit token: device ID not found")
            return null
        }

        val accessToken = getAccessToken() ?: run {
            Log.e(TAG, "Cannot get LiveKit token: access token not found")
            return null
        }

        return withContext(Dispatchers.IO) {
            try {
                val server = getServerUrl()
                val url = "$server/api/v1/device/$deviceId/livekit-token"

                val req = Request.Builder()
                    .url(url)
                    .get()
                    .addHeader("Authorization", "Bearer $accessToken")
                    .build()

                client.newCall(req).execute().use { resp ->
                    if (resp.isSuccessful) {
                        val text = resp.body?.string() ?: ""
                        Log.d(TAG, "LiveKit token response received")

                        val json = JSONObject(text)
                        val data = json.getJSONObject("data")
                        val token = data.getString("token")
                        val roomId = data.getString("room_id")
                        val serverUrl = data.getString("server_url")

                        Log.d(TAG, "LiveKit token fetched successfully")
                        Log.d(TAG, "- Room ID: $roomId")
                        Log.d(TAG, "- Server URL: $serverUrl")

                        Triple(token, roomId, serverUrl)
                    } else {
                        val errorText = resp.body?.string() ?: "Unknown error"
                        Log.e(TAG, "Failed to get LiveKit token: ${resp.code} - $errorText")
                        null
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Error getting LiveKit token", e)
                null
            }
        }
    }
}

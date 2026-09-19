package com.rvcs.android

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.util.Log
import android.widget.Button
import android.widget.EditText
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import com.google.gson.Gson
import com.google.gson.JsonObject
import kotlinx.coroutines.*
import okhttp3.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException

class RegistrationCodeActivity : AppCompatActivity() {
    companion object {
        private const val TAG = "RegistrationCode"
        private const val CAMERA_PERMISSION_REQUEST = 1001
        private const val DEFAULT_SERVER_URL = "https://YOUR_SERVER_IP:28443" // 默认服务端地址
        private const val PREFS_NAME = "device_prefs"
        private const val KEY_SERVER_URL = "server_url"
    }

    private lateinit var serverUrlEditText: EditText
    private lateinit var registrationCodeEditText: EditText
    private lateinit var deviceNameEditText: EditText
    private lateinit var registerButton: Button
    private lateinit var statusTextView: TextView
    private lateinit var goToMainButton: Button
    
    private val gson = Gson()
    private val client = SSLUtils.createUnsafeOkHttpClient()
    private val scope = CoroutineScope(Dispatchers.Main + Job())

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_registration_code)
        
        initViews()
        setupClickListeners()
        
        // 检查必要权限
        checkPermissions()
    }

    private fun initViews() {
        serverUrlEditText = findViewById(R.id.et_server_url)
        registrationCodeEditText = findViewById(R.id.et_registration_code)
        deviceNameEditText = findViewById(R.id.et_device_name)
        registerButton = findViewById(R.id.btn_register)
        statusTextView = findViewById(R.id.tv_status)
        goToMainButton = findViewById(R.id.btn_go_to_main)
        
        // 加载保存的服务端地址或使用默认值
        val savedServerUrl = getSavedServerUrl()
        serverUrlEditText.setText(savedServerUrl)
        
        // 默认设备名称
        deviceNameEditText.setText("Android_Camera_Device_${android.os.Build.MODEL}")
    }

    private fun setupClickListeners() {
        registerButton.setOnClickListener {
            registerDeviceWithCode()
        }
        
        goToMainButton.setOnClickListener {
            startActivity(Intent(this, MainActivity::class.java))
            finish()
        }
    }

    private fun checkPermissions() {
        val permissions = arrayOf(
            Manifest.permission.CAMERA,
            Manifest.permission.INTERNET,
            Manifest.permission.ACCESS_NETWORK_STATE
        )
        
        val permissionsToRequest = permissions.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }
        
        if (permissionsToRequest.isNotEmpty()) {
            ActivityCompat.requestPermissions(
                this,
                permissionsToRequest.toTypedArray(),
                CAMERA_PERMISSION_REQUEST
            )
        }
    }

    private fun registerDeviceWithCode() {
        val serverUrl = serverUrlEditText.text.toString().trim()
        val registrationCode = registrationCodeEditText.text.toString().trim()
        val deviceName = deviceNameEditText.text.toString().trim()
        
        if (serverUrl.isEmpty()) {
            Toast.makeText(this, "请输入服务端地址", Toast.LENGTH_SHORT).show()
            return
        }
        
        if (registrationCode.isEmpty()) {
            Toast.makeText(this, "请输入注册码", Toast.LENGTH_SHORT).show()
            return
        }
        
        if (deviceName.isEmpty()) {
            Toast.makeText(this, "请输入设备名称", Toast.LENGTH_SHORT).show()
            return
        }
        
        // 验证服务端地址格式
        if (!isValidServerUrl(serverUrl)) {
            Toast.makeText(this, "服务端地址格式不正确，请使用 https://ip:port 格式", Toast.LENGTH_LONG).show()
            return
        }
        
        // 验证注册码格式
        if (!isValidRegistrationCode(registrationCode)) {
            Toast.makeText(this, "注册码格式不正确", Toast.LENGTH_SHORT).show()
            return
        }
        
        registerButton.isEnabled = false
        statusTextView.text = "正在连接服务端..."
        
        scope.launch {
            try {
                val registrationResult = performRegistrationWithCode(serverUrl, registrationCode, deviceName)
                if (registrationResult.success) {
                    statusTextView.text = "✅ 设备注册成功！"
                    registerButton.text = "注册成功"
                    goToMainButton.isEnabled = true
                    
                    // 保存服务端地址和注册信息
                    saveServerUrl(serverUrl)
                    saveRegistrationInfo(registrationResult.deviceId, registrationResult.deviceToken!!, registrationResult.refreshToken!!, deviceName)
                    
                    Toast.makeText(this@RegistrationCodeActivity, "设备注册成功！", Toast.LENGTH_LONG).show()
                } else {
                    statusTextView.text = "❌ 注册失败: ${registrationResult.errorMessage}"
                    registerButton.isEnabled = true
                }
            } catch (e: Exception) {
                Log.e(TAG, "Registration failed", e)
                statusTextView.text = "❌ 注册异常: ${e.message}"
                registerButton.isEnabled = true
            }
        }
    }

    private fun isValidRegistrationCode(code: String): Boolean {
        // 简单的格式验证：至少包含"RVCS-"前缀和足够的长度
        return code.startsWith("RVCS-") && code.length >= 32
    }
    
    private fun isValidServerUrl(url: String): Boolean {
        return try {
            val uri = java.net.URI(url)
            uri.scheme == "https" && uri.host != null && uri.port > 0
        } catch (e: Exception) {
            false
        }
    }
    
    private fun getSavedServerUrl(): String {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        return prefs.getString(KEY_SERVER_URL, DEFAULT_SERVER_URL) ?: DEFAULT_SERVER_URL
    }
    
    private fun saveServerUrl(serverUrl: String) {
        val prefs = getSharedPreferences(PREFS_NAME, MODE_PRIVATE)
        with(prefs.edit()) {
            putString(KEY_SERVER_URL, serverUrl)
            apply()
        }
    }

    private suspend fun performRegistrationWithCode(serverUrl: String, registrationCode: String, deviceName: String): RegistrationResult = withContext(Dispatchers.IO) {
        try {
            val json = JsonObject().apply {
                addProperty("registration_code", registrationCode)
                addProperty("device_name", deviceName)
                addProperty("manufacturer", android.os.Build.MANUFACTURER)
                addProperty("android_version", android.os.Build.VERSION.RELEASE)
                addProperty("app_version", "1.0.0")
            }.toString()
            
            val body = json.toRequestBody("application/json".toMediaType())
            val request = Request.Builder()
                .url("$serverUrl/api/v1/device/register-with-code")
                .post(body)
                .addHeader("Content-Type", "application/json")
                .build()
            
            client.newCall(request).execute().use { response ->
                Log.d(TAG, "Registration response code: ${response.code}")
                
                if (response.isSuccessful) {
                    val responseBody = response.body?.string() ?: ""
                    Log.d(TAG, "Registration response: $responseBody")
                    
                    try {
                        val responseObject = gson.fromJson(responseBody, JsonObject::class.java)
                        val data = responseObject.getAsJsonObject("data")
                        
                        return@withContext RegistrationResult(
                            success = true,
                            deviceToken = data.get("device_token").asString,
                            refreshToken = data.get("refresh_token").asString,
                            deviceId = data.get("device_id").asString
                        )
                    } catch (e: Exception) {
                        Log.e(TAG, "Failed to parse registration response", e)
                        return@withContext RegistrationResult(
                            success = false,
                            errorMessage = "响应数据解析失败"
                        )
                    }
                } else {
                    val errorBody = response.body?.string() ?: ""
                    Log.e(TAG, "Registration failed: ${response.code} - $errorBody")
                    
                    try {
                        val errorObject = gson.fromJson(errorBody, JsonObject::class.java)
                        val errorMessage = errorObject.get("message")?.asString ?: "注册失败"
                        return@withContext RegistrationResult(
                            success = false,
                            errorMessage = errorMessage
                        )
                    } catch (e: Exception) {
                        return@withContext RegistrationResult(
                            success = false,
                            errorMessage = "HTTP ${response.code}: $errorBody"
                        )
                    }
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Network error during registration", e)
            return@withContext RegistrationResult(
                success = false,
                errorMessage = "网络连接错误: ${e.message}"
            )
        }
    }

    private fun saveRegistrationInfo(deviceId: String?, deviceToken: String, refreshToken: String, deviceName: String) {
        val prefs = getSharedPreferences("device_prefs", MODE_PRIVATE)
        with(prefs.edit()) {
            if (deviceId != null) {
                putString("device_id", deviceId)
            }
            putString("device_token", deviceToken)
            putString("refresh_token", refreshToken)
            putString("device_name", deviceName)
            putBoolean("is_registered", true)
            apply()
        }
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        
        if (requestCode == CAMERA_PERMISSION_REQUEST) {
            if (grantResults.isNotEmpty() && grantResults.all { it == PackageManager.PERMISSION_GRANTED }) {
                Toast.makeText(this, "权限获取成功", Toast.LENGTH_SHORT).show()
            } else {
                Toast.makeText(this, "需要相机权限才能正常使用", Toast.LENGTH_LONG).show()
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        scope.cancel()
    }

    // 注册结果数据类
    private data class RegistrationResult(
        val success: Boolean,
        val deviceToken: String? = null,
        val refreshToken: String? = null,
        val deviceId: String? = null,
        val errorMessage: String? = null
    )
}

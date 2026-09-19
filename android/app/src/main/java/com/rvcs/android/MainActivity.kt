package com.rvcs.android

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.util.Log
import android.widget.Button
import android.widget.TextView
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.lifecycle.lifecycleScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class MainActivity : AppCompatActivity() {
    private lateinit var statusText: TextView
    private lateinit var startBtn: Button
    private lateinit var stopBtn: Button
    private lateinit var deviceIdText: TextView
    private lateinit var tokenText: TextView
    private lateinit var batteryOptText: TextView
    private lateinit var settingsBtn: Button
    private lateinit var reRegisterBtn: Button
    private lateinit var debugPrefsBtn: Button

    private var isServiceRunning = false
    private var batteryOptHelper: BatteryOptimizationHelper? = null

    private val REQUEST_IGNORE_BATTERY_OPTIMIZATIONS = 1001
    private val REQUEST_AUDIO_PERMISSION = 1002

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        batteryOptHelper = BatteryOptimizationHelper(this)

        // 检查设备是否已注册
        checkDeviceRegistration()

        initViews()
        checkServiceStatus()
        checkBatteryOptimization()
    }

    private fun initViews() {
        statusText = findViewById(R.id.tv_status)
        startBtn = findViewById(R.id.btn_start_service)
        stopBtn = findViewById(R.id.btn_stop_service)
        deviceIdText = findViewById(R.id.tv_device_id)
        tokenText = findViewById(R.id.tv_token)
        reRegisterBtn = findViewById(R.id.btn_re_register)
        val openCameraPreviewBtn = findViewById<Button>(R.id.btn_open_camera_preview)

        // 查找或创建电池优化状态视图（如果布局中没有，可以忽略）
        try {
            batteryOptText = findViewById(R.id.tv_battery_opt)
            settingsBtn = findViewById(R.id.btn_settings)
            settingsBtn.setOnClickListener {
                openSettings()
            }
        } catch (e: Exception) {
            // 如果布局中没有这些视图，就忽略
        }

        // 查找调试按钮并设置点击事件
        debugPrefsBtn = findViewById(R.id.btn_debug_prefs)
        debugPrefsBtn.setOnClickListener {
            Log.d("MainActivity", "DebugPrefs button clicked")
            val intent = Intent(this, DebugPrefsActivity::class.java)
            startActivity(intent)
        }

        startBtn.setOnClickListener {
            startCameraService()
        }

        stopBtn.setOnClickListener {
            stopCameraService()
        }

        // 暂时禁用相机预览按钮（功能正在开发中）
        openCameraPreviewBtn.setOnClickListener {
            Log.d("MainActivity", "Camera preview feature is disabled")
        }

        updateUI(false)
        updateDeviceInfo() // 初始化时更新设备信息显示
    }
    
    /**
     * 检查设备注册状态
     */
    private fun checkDeviceRegistration() {
        Log.d("MainActivity", "=== 开始检查设备注册状态 ===")
        val registrar = DeviceRegistrar(this)
        val isRegistered = registrar.isRegistered()

        Log.d("MainActivity", "Device registered status: $isRegistered")
        Log.d("MainActivity", "Device ID: ${registrar.getDeviceId()}")
        Log.d("MainActivity", "Access Token: ${registrar.getAccessToken()?.take(20)}...")
        
        // 详细检查各个条件
        val deviceId = registrar.getDeviceId()
        val accessToken = registrar.getAccessToken()
        val isRegisteredFlag = getSharedPreferences("device_prefs", MODE_PRIVATE).getBoolean("is_registered", false)
        
        Log.d("MainActivity", "详细检查:")
        Log.d("MainActivity", "- deviceId存在: ${deviceId != null}")
        Log.d("MainActivity", "- accessToken存在: ${accessToken != null}")
        Log.d("MainActivity", "- is_registered标志: $isRegisteredFlag")
        Log.d("MainActivity", "- 最终结果: $isRegistered")

        if (!isRegistered) {
            Log.d("MainActivity", "设备未注册，准备跳转到注册页面")
            // 跳转到设备注册页面
            val intent = Intent(this, RegistrationCodeActivity::class.java)
            startActivity(intent)
            finish()
            return
        } else {
            Log.d("MainActivity", "设备已注册，继续启动主界面")
        }
        Log.d("MainActivity", "=== 设备注册状态检查结束 ===")
    }
    
    private fun checkBatteryOptimization() {
        val helper = batteryOptHelper ?: return
        
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            if (!helper.isIgnoringBatteryOptimizations()) {
                // 请求忽略电池优化
                val intent = helper.requestIgnoreBatteryOptimizations()
                if (intent != null) {
                    try {
                        startActivityForResult(intent, REQUEST_IGNORE_BATTERY_OPTIMIZATIONS)
                    } catch (e: Exception) {
                        // 用户可能已经拒绝或取消
                        Log.w("MainActivity", "Failed to request battery optimization permission", e)
                    }
                }
            }
        }
        
        // 更新显示
        updateBatteryOptimizationStatus()
    }
    
    /**
     * 更新电池优化状态显示
     */
    private fun updateBatteryOptimizationStatus() {
        val helper = batteryOptHelper ?: return
        val tips = helper.getOptimizationTips()
        
        if (tips.isNotEmpty() && ::batteryOptText.isInitialized) {
            batteryOptText.text = "建议: ${tips[0]}"
        } else if (::batteryOptText.isInitialized) {
            batteryOptText.text = "电池优化: 已优化"
        }
    }
    
    /**
     * 打开设置页面
     */
    private fun openSettings() {
        val helper = batteryOptHelper ?: return
        val intent = helper.openBatteryOptimizationSettings()
        startActivity(intent)
    }
    
    /**
     * 处理权限请求结果
     */
    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)

        when (requestCode) {
            REQUEST_IGNORE_BATTERY_OPTIMIZATIONS -> {
                updateBatteryOptimizationStatus()
                batteryOptHelper?.logOptimizationStatus()
            }
        }
    }

    /**
     * 处理运行时权限请求结果
     */
    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)

        when (requestCode) {
            REQUEST_AUDIO_PERMISSION -> {
                if (grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                    Toast.makeText(this, "录音权限已授予", Toast.LENGTH_SHORT).show()
                    // 重新启动服务
                    startCameraService()
                } else {
                    Toast.makeText(this, "需要录音权限才能推送音频", Toast.LENGTH_LONG).show()
                }
            }
        }
    }

    private fun checkServiceStatus() {
        lifecycleScope.launch {
            while (true) {
                try {
                    updateDeviceInfo()
                    // 降低频率到30分钟，因为已有心跳机制
                    kotlinx.coroutines.delay(30 * 60 * 1000L) // 30分钟
                } catch (e: Exception) {
                    Log.e("MainActivity", "Error in checkServiceStatus", e)
                    kotlinx.coroutines.delay(30 * 60 * 1000L) // 30分钟
                }
            }
        }
    }

    private fun startCameraService() {
        if (isServiceRunning) {
            return
        }

        // 检查录音权限
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.RECORD_AUDIO)
            != PackageManager.PERMISSION_GRANTED) {
            // 请求权限
            ActivityCompat.requestPermissions(
                this,
                arrayOf(Manifest.permission.RECORD_AUDIO),
                REQUEST_AUDIO_PERMISSION
            )
            return
        }

        val intent = Intent(this, ForegroundCameraService::class.java).apply {
            action = ForegroundCameraService.ACTION_START
        }

        // 兼容API级别处理
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            startForegroundService(intent)
        } else {
            startService(intent)
        }

        isServiceRunning = true
        updateUI(true)
        statusText.text = "Starting service..."
    }

    private fun stopCameraService() {
        if (!isServiceRunning) {
            return
        }

        val intent = Intent(this, ForegroundCameraService::class.java).apply {
            action = ForegroundCameraService.ACTION_STOP
        }
        startService(intent)

        isServiceRunning = false
        updateUI(false)
        statusText.text = "Service stopped"
    }

    private fun updateDeviceInfo() {
        val registrar = DeviceRegistrar(this)
        val deviceId = registrar.getDeviceId()
        val token = registrar.getAccessToken()
        val isRegistered = registrar.isRegistered()

        // 根据注册状态显示设备ID
        deviceIdText.text = if (isRegistered) {
            deviceId ?: "Registered (No Device ID)"
        } else {
            "Not registered"
        }
        
        tokenText.text = if (token != null) "Available (${token.take(20)}...)" else "Not available"

        // 更新状态文本
        when {
            !isRegistered -> {
                statusText.text = "Not registered - Please register your device first"
            }
            isServiceRunning -> {
                statusText.text = "✅ Service running - Device: ${deviceId ?: "Unknown"}"
            }
            else -> {
                statusText.text = "⏸️ Service stopped - Device: ${deviceId ?: "Unknown"}"
            }
        }
    }

    private fun updateUI(running: Boolean) {
        startBtn.isEnabled = !running
        stopBtn.isEnabled = running
    }

    override fun onDestroy() {
        super.onDestroy()
        // 相机资源由CameraPreviewActivity管理
    }
}
package com.rvcs.android

import android.os.Bundle
import android.util.Log
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class DebugActivity : AppCompatActivity() {
    companion object {
        private const val TAG = "DebugActivity"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        val layout = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.VERTICAL
            setPadding(50, 50, 50, 50)
        }
        
        val textView = TextView(this).apply {
            text = "调试信息正在收集..."
            textSize = 16f
        }
        
        val clearButton = Button(this).apply {
            text = "清除注册数据"
            setOnClickListener {
                clearRegistrationData()
                updateDebugInfo(textView)
            }
        }
        
        val refreshButton = Button(this).apply {
            text = "刷新信息"
            setOnClickListener {
                updateDebugInfo(textView)
            }
        }
        
        layout.addView(textView)
        layout.addView(clearButton)
        layout.addView(refreshButton)
        
        setContentView(layout)
        
        updateDebugInfo(textView)
    }
    
    private fun updateDebugInfo(textView: TextView) {
        val registrar = DeviceRegistrar(this)
        val cleaner = DeviceRegistrationCleaner(this)
        
        val status = cleaner.checkRegistrationStatus()
        val deviceId = registrar.getDeviceId()
        val accessToken = registrar.getAccessToken()
        val isRegistered = registrar.isRegistered()
        
        val debugInfo = buildString {
            appendLine("=== 设备注册状态调试 ===")
            appendLine("Device ID: $deviceId")
            appendLine("Access Token: ${if (accessToken != null) "存在(${accessToken.take(20)}...)" else "不存在"}")
            appendLine("isRegistered(): $isRegistered")
            appendLine("")
            appendLine("详细状态:")
            status.forEach { (key, value) ->
                appendLine("- $key: $value")
            }
            appendLine("")
            appendLine("注册判断逻辑:")
            appendLine("- deviceId != null: ${deviceId != null}")
            appendLine("- accessToken != null: ${accessToken != null}")
            appendLine("- is_registered标志: ${getSharedPreferences("device_prefs", MODE_PRIVATE).getBoolean("is_registered", false)}")
            appendLine("- 综合判断结果: ${(deviceId != null && accessToken != null) || getSharedPreferences("device_prefs", MODE_PRIVATE).getBoolean("is_registered", false)}")
        }
        
        textView.text = debugInfo
        Log.d(TAG, debugInfo)
    }
    
    private fun clearRegistrationData() {
        val cleaner = DeviceRegistrationCleaner(this)
        cleaner.clearRegistration()
        Log.d(TAG, "注册数据已清除")
    }
}
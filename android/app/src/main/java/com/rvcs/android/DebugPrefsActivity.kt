package com.rvcs.android

import android.content.Context
import android.content.SharedPreferences
import android.os.Bundle
import android.util.Log
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class DebugPrefsActivity : AppCompatActivity() {
    companion object {
        private const val TAG = "DebugPrefsActivity"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        val layout = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.VERTICAL
            setPadding(50, 50, 50, 50)
        }
        
        val textView = TextView(this).apply {
            text = "正在检查SharedPreferences数据..."
            textSize = 14f
        }
        
        val refreshButton = Button(this).apply {
            text = "刷新数据"
            setOnClickListener {
                updatePrefsInfo(textView)
            }
        }
        
        val clearButton = Button(this).apply {
            text = "清除所有数据"
            setOnClickListener {
                clearAllPrefs()
                updatePrefsInfo(textView)
            }
        }
        
        layout.addView(textView)
        layout.addView(refreshButton)
        layout.addView(clearButton)
        
        setContentView(layout)
        
        updatePrefsInfo(textView)
    }
    
    private fun updatePrefsInfo(textView: TextView) {
        val prefs = getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        
        val deviceId = prefs.getString("device_id", null)
        val deviceToken = prefs.getString("device_token", null)
        val refreshToken = prefs.getString("refresh_token", null)
        val isRegistered = prefs.getBoolean("is_registered", false)
        val serverUrl = prefs.getString("server_url", null)
        
        val debugInfo = buildString {
            append("=== SharedPreferences 调试信息 ===\n\n")
            
            append("device_id: ${if (deviceId != null) deviceId else "NULL"}\n")
            append("device_token: ${if (deviceToken != null) "${deviceToken.take(20)}..." else "NULL"}\n")
            append("refresh_token: ${if (refreshToken != null) "${refreshToken.take(20)}..." else "NULL"}\n")
            append("is_registered: $isRegistered\n")
            append("server_url: ${serverUrl ?: "NULL"}\n\n")
            
            append("=== 计算结果 ===\n")
            append("deviceId != null: ${deviceId != null}\n")
            append("deviceToken != null: ${deviceToken != null}\n")
            append("isRegistered flag: $isRegistered\n")
            append("最终 isRegistered 结果: ${(deviceId != null && deviceToken != null) || isRegistered}\n")
        }
        
        textView.text = debugInfo
        Log.d(TAG, debugInfo)
    }
    
    private fun clearAllPrefs() {
        val prefs = getSharedPreferences("device_prefs", Context.MODE_PRIVATE)
        prefs.edit().clear().apply()
        Log.d(TAG, "All preferences cleared")
    }
}
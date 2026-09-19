package com.rvcs.android

import android.util.Log
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

/**
 * 质量控制器
 * 负责根据网络状况自动调整视频流质量
 */
class QualityController {
    private val TAG = "QualityController"
    
    // 网络状况枚举
    enum class NetworkCondition {
        EXCELLENT,    // 优秀 (>5Mbps)
        GOOD,         // 良好 (2-5Mbps)
        FAIR,         // 一般 (1-2Mbps)
        POOR,         // 较差 (<1Mbps)
        VERY_POOR     // 很差 (<500Kbps)
    }
    
    // 质量配置
    data class QualityProfile(
        val name: String,
        val width: Int,
        val height: Int,
        val bitrate: Int,
        val fps: Int,
        val iFrameInterval: Int = 1
    ) {
        companion object {
            val LOW = QualityProfile("low", 640, 480, 800000, 15)
            val MEDIUM = QualityProfile("medium", 1280, 720, 2000000, 30)
            val HIGH = QualityProfile("high", 1920, 1080, 4000000, 30)
            val AUTO = QualityProfile("auto", 1280, 720, 2000000, 30) // 自适应模式
            
            fun getAllProfiles(): List<QualityProfile> = listOf(LOW, MEDIUM, HIGH, AUTO)
            fun getDefault(): QualityProfile = MEDIUM
        }
    }
    
    // 网络监测器接口
    interface NetworkMonitor {
        suspend fun getBandwidth(): Long // 返回bps
        suspend fun getPacketLossRate(): Float // 返回丢包率(0-1)
        suspend fun getLatency(): Long // 返回延迟(ms)
    }
    
    private val _currentQuality = MutableStateFlow(QualityProfile.getDefault())
    val currentQuality = _currentQuality.asStateFlow()
    
    private val _networkCondition = MutableStateFlow(NetworkCondition.FAIR)
    val networkCondition = _networkCondition.asStateFlow()
    
    private var networkMonitor: NetworkMonitor? = null
    private var autoAdjustEnabled = false
    private var monitoringJob: Job? = null
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    
    // 调整阈值
    private val bandwidthThresholds = mapOf(
        NetworkCondition.EXCELLENT to 5000000L, // 5Mbps
        NetworkCondition.GOOD to 2000000L,      // 2Mbps
        NetworkCondition.FAIR to 1000000L,      // 1Mbps
        NetworkCondition.POOR to 500000L,       // 500Kbps
        NetworkCondition.VERY_POOR to 0L
    )
    
    /**
     * 设置网络监测器
     */
    fun setNetworkMonitor(monitor: NetworkMonitor) {
        this.networkMonitor = monitor
    }
    
    /**
     * 启用自动调整
     */
    fun enableAutoAdjust(checkIntervalMs: Long = 5000L) {
        if (autoAdjustEnabled) return
        
        autoAdjustEnabled = true
        monitoringJob = scope.launch {
            while (isActive && autoAdjustEnabled) {
                try {
                    adjustQualityIfNeeded()
                    delay(checkIntervalMs)
                } catch (e: Exception) {
                    Log.e(TAG, "Auto adjustment error", e)
                }
            }
        }
        Log.d(TAG, "Auto quality adjustment enabled")
    }
    
    /**
     * 禁用自动调整
     */
    fun disableAutoAdjust() {
        autoAdjustEnabled = false
        monitoringJob?.cancel()
        monitoringJob = null
        Log.d(TAG, "Auto quality adjustment disabled")
    }
    
    /**
     * 手动设置质量
     */
    fun setQuality(profile: QualityProfile) {
        if (profile.name == "auto") {
            enableAutoAdjust()
        } else {
            disableAutoAdjust()
            _currentQuality.value = profile
            Log.d(TAG, "Manual quality set to: ${profile.name}")
        }
    }
    
    /**
     * 根据网络状况调整质量
     */
    private suspend fun adjustQualityIfNeeded() {
        val monitor = networkMonitor ?: return
        val bandwidth = monitor.getBandwidth()
        val packetLoss = monitor.getPacketLossRate()
        val latency = monitor.getLatency()
        
        val newCondition = evaluateNetworkCondition(bandwidth, packetLoss, latency)
        val newQuality = suggestQuality(newCondition)
        
        // 只有在网络条件显著变化时才调整
        if (newCondition != _networkCondition.value || 
            shouldSwitchQuality(_currentQuality.value, newQuality)) {
            
            _networkCondition.value = newCondition
            _currentQuality.value = newQuality
            
            Log.d(TAG, "Quality adjusted: ${newQuality.name} (${newCondition})")
            Log.d(TAG, "Bandwidth: ${bandwidth}bps, PacketLoss: ${packetLoss*100}%, Latency: ${latency}ms")
        }
    }
    
    /**
     * 评估网络状况
     */
    private fun evaluateNetworkCondition(bandwidth: Long, packetLoss: Float, latency: Long): NetworkCondition {
        return when {
            bandwidth >= bandwidthThresholds[NetworkCondition.EXCELLENT]!! && 
            packetLoss < 0.02f && latency < 100 -> NetworkCondition.EXCELLENT
            
            bandwidth >= bandwidthThresholds[NetworkCondition.GOOD]!! && 
            packetLoss < 0.05f && latency < 200 -> NetworkCondition.GOOD
            
            bandwidth >= bandwidthThresholds[NetworkCondition.FAIR]!! && 
            packetLoss < 0.1f && latency < 300 -> NetworkCondition.FAIR
            
            bandwidth >= bandwidthThresholds[NetworkCondition.POOR]!! -> NetworkCondition.POOR
            
            else -> NetworkCondition.VERY_POOR
        }
    }
    
    /**
     * 根据网络状况建议质量
     */
    private fun suggestQuality(condition: NetworkCondition): QualityProfile {
        return when (condition) {
            NetworkCondition.EXCELLENT -> QualityProfile.HIGH
            NetworkCondition.GOOD -> QualityProfile.HIGH
            NetworkCondition.FAIR -> QualityProfile.MEDIUM
            NetworkCondition.POOR -> QualityProfile.LOW
            NetworkCondition.VERY_POOR -> QualityProfile.LOW
        }
    }
    
    /**
     * 判断是否需要切换质量
     */
    private fun shouldSwitchQuality(current: QualityProfile, suggested: QualityProfile): Boolean {
        // 如果已经是建议质量，不需要切换
        if (current.name == suggested.name) return false
        
        // 避免频繁切换，只在质量差异较大时切换
        val currentBandwidth = when (current.name) {
            "low" -> 1000000
            "medium" -> 2000000
            "high" -> 4000000
            else -> 2000000
        }
        
        val suggestedBandwidth = when (suggested.name) {
            "low" -> 1000000
            "medium" -> 2000000
            "high" -> 4000000
            else -> 2000000
        }
        
        // 只有当带宽需求差异超过50%时才考虑切换
        val diffRatio = kotlin.math.abs(currentBandwidth - suggestedBandwidth).toFloat() / 
                       kotlin.math.max(currentBandwidth, suggestedBandwidth).toFloat()
        
        return diffRatio > 0.5f
    }
    
    /**
     * 获取当前网络统计信息
     */
    suspend fun getNetworkStats(): Map<String, Any> {
        val monitor = networkMonitor ?: return emptyMap()
        return mapOf(
            "bandwidth" to monitor.getBandwidth(),
            "packetLoss" to monitor.getPacketLossRate(),
            "latency" to monitor.getLatency(),
            "condition" to _networkCondition.value.name,
            "quality" to _currentQuality.value.name
        )
    }
    
    /**
     * 简单的网络监测器实现（基于系统API）
     */
    class SimpleNetworkMonitor : NetworkMonitor {
        override suspend fun getBandwidth(): Long {
            // 简化实现：模拟不同的网络状况
            // 实际应用中应该使用TrafficStats或ConnectivityManager
            return when ((0..100).random()) {
                in 0..10 -> 500000L    // 500Kbps - 很差
                in 11..30 -> 1500000L  // 1.5Mbps - 较差
                in 31..60 -> 2500000L  // 2.5Mbps - 一般
                in 61..85 -> 3500000L  // 3.5Mbps - 良好
                else -> 6000000L       // 6Mbps - 优秀
            }
        }
        
        override suspend fun getPacketLossRate(): Float {
            return when ((0..100).random()) {
                in 0..5 -> 0.01f   // 1% 丢包
                in 6..20 -> 0.05f  // 5% 丢包
                in 21..50 -> 0.1f  // 10% 丢包
                else -> 0.2f       // 20% 丢包
            }
        }
        
        override suspend fun getLatency(): Long {
            return when ((0..100).random()) {
                in 0..20 -> 50L    // 50ms 延迟
                in 21..50 -> 150L  // 150ms 延迟
                in 51..80 -> 300L  // 300ms 延迟
                else -> 500L       // 500ms 延迟
            }
        }
    }
    
    /**
     * 释放资源
     */
    fun destroy() {
        disableAutoAdjust()
        scope.cancel()
        Log.d(TAG, "QualityController destroyed")
    }
}
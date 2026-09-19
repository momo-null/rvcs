package com.rvcs.android

import android.Manifest
import android.content.Context
import android.content.pm.PackageManager
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import android.os.Build
import android.telephony.TelephonyManager
import android.util.Log
import androidx.annotation.RequiresApi
import androidx.core.content.ContextCompat

/**
 * 网络监控器
 * 监控网络类型、信号强度、网络质量
 */
class NetworkMonitor(private val context: Context) {

    private val TAG = "NetworkMonitor"

    private val connectivityManager: ConnectivityManager =
        context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private val telephonyManager: TelephonyManager =
        context.getSystemService(Context.TELEPHONY_SERVICE) as TelephonyManager

    private var networkType = NetworkType.UNKNOWN
    private var signalStrength = -100 // dBm
    private var isNetworkAvailable = false

    private var onNetworkChangeListener: ((NetworkInfo) -> Unit)? = null

    enum class NetworkType {
        UNKNOWN,
        WIFI,
        CELLULAR,
        ETHERNET,
        VPN,
        NONE
    }

    data class NetworkInfo(
        val type: NetworkType,
        val signalStrength: Int, // dBm
        val isAvailable: Boolean,
        val quality: NetworkQuality
    )

    enum class NetworkQuality {
        EXCELLENT, // > -70 dBm
        GOOD,       // -70 to -85 dBm
        FAIR,       // -86 to -95 dBm
        POOR        // < -95 dBm
    }

    /**
     * 获取当前网络类型
     */
    fun getCurrentNetworkType(): NetworkType {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            val network = connectivityManager.activeNetwork ?: return NetworkType.NONE
            val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return NetworkType.NONE

            return when {
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_WIFI) -> {
                    // 检查是否为 VPN
                    if (capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN)) {
                        NetworkType.VPN
                    } else {
                        NetworkType.WIFI
                    }
                }
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR) -> {
                    if (capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN)) {
                        NetworkType.VPN
                    } else {
                        NetworkType.CELLULAR
                    }
                }
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET) -> NetworkType.ETHERNET
                capabilities.hasTransport(NetworkCapabilities.TRANSPORT_VPN) -> NetworkType.VPN
                else -> NetworkType.UNKNOWN
            }
        } else {
            @Suppress("DEPRECATION")
            val activeNetworkInfo = connectivityManager.activeNetworkInfo ?: return NetworkType.NONE
            @Suppress("DEPRECATION")
            return when (activeNetworkInfo.type) {
                ConnectivityManager.TYPE_WIFI -> NetworkType.WIFI
                ConnectivityManager.TYPE_MOBILE -> NetworkType.CELLULAR
                ConnectivityManager.TYPE_ETHERNET -> NetworkType.ETHERNET
                else -> NetworkType.UNKNOWN
            }
        }
    }

    /**
     * 获取网络类型字符串
     */
    fun getNetworkTypeString(): String {
        return when (getCurrentNetworkType()) {
            NetworkType.WIFI -> "wifi"
            NetworkType.CELLULAR -> "cellular"
            NetworkType.ETHERNET -> "ethernet"
            NetworkType.VPN -> "vpn"
            NetworkType.NONE -> "none"
            NetworkType.UNKNOWN -> "unknown"
        }
    }

    /**
     * 获取信号强度
     */
    fun getSignalStrength(): Int {
        val networkType = getCurrentNetworkType()

        return when (networkType) {
            NetworkType.WIFI -> getWifiSignalStrength()
            NetworkType.CELLULAR -> getCellularSignalStrength()
            else -> -100
        }
    }

    /**
     * 获取 WiFi 信号强度
     */
    private fun getWifiSignalStrength(): Int {
        // Android 不直接提供 WiFi 信号强度的 API
        // 这里返回一个估算值
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            // Android 11+ 可以从 WifiManager 获取更多信息
            try {
                val wifiManager = context.getSystemService(Context.WIFI_SERVICE) as android.net.wifi.WifiManager
                val wifiInfo = wifiManager.connectionInfo
                // RSSI 通常在 -50 到 -100 之间
                wifiInfo.rssi
            } catch (e: Exception) {
                Log.e(TAG, "Failed to get WiFi signal strength", e)
                -70 // 默认值
            }
        } else {
            -70 // 默认值
        }
    }

    /**
     * 获取蜂窝网络信号强度
     */
    @Suppress("DEPRECATION")
    private fun getCellularSignalStrength(): Int {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            // Android 10+ 使用新 API
            try {
                val signalStrengths = telephonyManager.allCellInfo
                var bestSignal = -100

                signalStrengths?.forEach { cellInfo ->
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.JELLY_BEAN_MR1) {
                        when (cellInfo) {
                            is android.telephony.CellInfoGsm -> {
                                val strength = cellInfo.cellSignalStrength
                                if (strength != null) {
                                    bestSignal = maxOf(bestSignal, strength.dbm)
                                }
                            }
                            is android.telephony.CellInfoLte -> {
                                val strength = cellInfo.cellSignalStrength
                                if (strength != null) {
                                    bestSignal = maxOf(bestSignal, strength.dbm)
                                }
                            }
                            is android.telephony.CellInfoNr -> { // 5G
                                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                                    val strength = cellInfo.cellSignalStrength
                                    if (strength != null) {
                                        bestSignal = maxOf(bestSignal, strength.dbm)
                                    }
                                }
                            }
                        }
                    }
                }

                bestSignal
            } catch (e: Exception) {
                Log.e(TAG, "Failed to get cellular signal strength", e)
                -80 // 默认值
            }
        } else {
            // Android 10 以下使用已弃用的 API
            try {
                @Suppress("DEPRECATION")
                val signalStrength = telephonyManager.signalStrength
                if (signalStrength != null) {
                    @Suppress("DEPRECATION")
                    val gsmSignalStrength = signalStrength.gsmSignalStrength
                    // ASU 转换为 dBm: (ASU * 2) - 113
                    (gsmSignalStrength * 2) - 113
                } else {
                    -80
                }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to get cellular signal strength (legacy)", e)
                -80
            }
        }
    }

    /**
     * 获取网络质量
     */
    fun getNetworkQuality(): NetworkQuality {
        val signalStrength = getSignalStrength()
        return when {
            signalStrength > -70 -> NetworkQuality.EXCELLENT
            signalStrength > -85 -> NetworkQuality.GOOD
            signalStrength > -95 -> NetworkQuality.FAIR
            else -> NetworkQuality.POOR
        }
    }

    /**
     * 检查网络是否可用
     */
    fun isNetworkAvailable(): Boolean {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            val network = connectivityManager.activeNetwork ?: return false
            val capabilities = connectivityManager.getNetworkCapabilities(network) ?: return false
            return capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET) &&
                   capabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED)
        } else {
            @Suppress("DEPRECATION")
            val networkInfo = connectivityManager.activeNetworkInfo ?: return false
            @Suppress("DEPRECATION")
            return networkInfo.isConnected && networkInfo.isAvailable
        }
    }

    /**
     * 获取完整的网络信息
     */
    fun getNetworkInfo(): NetworkInfo {
        val type = getCurrentNetworkType()
        val signal = getSignalStrength()
        val available = isNetworkAvailable()
        val quality = getNetworkQuality()

        return NetworkInfo(type, signal, available, quality)
    }

    /**
     * 设置网络变化监听
     */
    @RequiresApi(Build.VERSION_CODES.N)
    fun registerNetworkCallback() {
        val networkRequest = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .build()

        val handler = android.os.Handler(android.os.Looper.getMainLooper())
        connectivityManager.registerDefaultNetworkCallback(
            object : ConnectivityManager.NetworkCallback() {
                override fun onAvailable(network: Network) {
                    super.onAvailable(network)
                    updateNetworkState()
                    Log.d(TAG, "Network available")
                }

                override fun onLost(network: Network) {
                    super.onLost(network)
                    updateNetworkState()
                    Log.d(TAG, "Network lost")
                }

                override fun onCapabilitiesChanged(
                    network: Network,
                    networkCapabilities: NetworkCapabilities
                ) {
                    super.onCapabilitiesChanged(network, networkCapabilities)
                    updateNetworkState()
                    Log.d(TAG, "Network capabilities changed")
                }
            },
            handler
        )
    }

    /**
     * 更新网络状态
     */
    private fun updateNetworkState() {
        networkType = getCurrentNetworkType()
        signalStrength = getSignalStrength()
        isNetworkAvailable = isNetworkAvailable()

        onNetworkChangeListener?.invoke(getNetworkInfo())
    }

    /**
     * 设置网络变化监听器
     */
    fun setOnNetworkChangeListener(listener: (NetworkInfo) -> Unit) {
        this.onNetworkChangeListener = listener
    }

    /**
     * 检查是否适合高质量视频流
     */
    fun isSuitableForHighQualityStream(): Boolean {
        val quality = getNetworkQuality()
        val type = getCurrentNetworkType()
        return quality != NetworkQuality.POOR &&
               (type == NetworkType.WIFI || quality == NetworkQuality.EXCELLENT)
    }

    /**
     * 获取推荐的流质量
     */
    fun getRecommendedStreamQuality(): String {
        return when (getNetworkQuality()) {
            NetworkQuality.EXCELLENT -> "high"
            NetworkQuality.GOOD -> "medium"
            NetworkQuality.FAIR -> "low"
            NetworkQuality.POOR -> "low"
        }
    }
}

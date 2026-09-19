package com.rvcs.android

/**
 * 视频分辨率配置
 */
enum class VideoResolution(
    val width: Int,
    val height: Int,
    val fps: Int,
    val minBitrate: Int,    // 最小码率 (kbps)
    val maxBitrate: Int,    // 最大码率 (kbps)
    val targetBitrate: Int, // 目标码率 (kbps)
    val displayName: String
) {
    RES_480P(640, 480, 30, 300, 800, 500, "480P"),
    RES_720P(1280, 720, 30, 500, 2000, 1200, "720P"),
    RES_1080P(1920, 1080, 30, 1500, 4000, 2500, "1080P");

    companion object {
        /**
         * 从字符串解析分辨率
         */
        fun fromString(str: String): VideoResolution {
            return values().find { it.displayName.equals(str, ignoreCase = true) } ?: RES_720P
        }

        /**
         * 根据网络质量自动选择分辨率
         */
        fun fromNetworkQuality(quality: NetworkQuality): VideoResolution {
            return when (quality) {
                NetworkQuality.POOR -> RES_480P
                NetworkQuality.GOOD -> RES_720P
                NetworkQuality.EXCELLENT -> RES_1080P
            }
        }
    }
}

/**
 * 网络质量等级
 */
enum class NetworkQuality(val level: Int, val displayName: String) {
    POOR(1, "差"),
    GOOD(2, "良好"),
    EXCELLENT(3, "优秀")
}

/**
 * 码率控制参数
 */
data class BitrateConfig(
    val minBitrate: Int,
    val maxBitrate: Int,
    val targetBitrate: Int
) {
    /**
     * 根据网络质量调整码率
     */
    fun adjustForNetworkQuality(quality: NetworkQuality): BitrateConfig {
        val multiplier = when (quality) {
            NetworkQuality.POOR -> 0.5
            NetworkQuality.GOOD -> 1.0
            NetworkQuality.EXCELLENT -> 1.5
        }
        return BitrateConfig(
            minBitrate = (minBitrate * multiplier).toInt().coerceAtLeast(100),
            maxBitrate = (maxBitrate * multiplier).toInt(),
            targetBitrate = (targetBitrate * multiplier).toInt()
        )
    }
}

package com.rvcs.android.network

import android.content.Context
import com.rvcs.android.BuildConfig
import com.rvcs.android.utils.CertificateUtils
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import java.util.concurrent.TimeUnit
import javax.net.ssl.SSLContext
import javax.net.ssl.X509TrustManager

/**
 * 网络客户端配置
 * 处理HTTPS自签证书和各种网络配置
 */
class NetworkClient private constructor() {
    
    companion object {
        @Volatile
        private var INSTANCE: OkHttpClient? = null
        
        fun getInstance(context: Context): OkHttpClient {
            return INSTANCE ?: synchronized(this) {
                INSTANCE ?: createOkHttpClient(context).also { INSTANCE = it }
            }
        }
        
        private fun createOkHttpClient(context: Context): OkHttpClient {
            val builder = OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(30, TimeUnit.SECONDS)
                .writeTimeout(30, TimeUnit.SECONDS)
            
            // 开发环境下配置自签证书
            if (BuildConfig.DEBUG) {
                configureDevelopmentCertificates(builder, context)
            }
            
            // 添加日志拦截器（仅开发环境）
            if (BuildConfig.DEBUG) {
                val loggingInterceptor = HttpLoggingInterceptor().apply {
                    level = HttpLoggingInterceptor.Level.BODY
                }
                builder.addInterceptor(loggingInterceptor)
            }
            
            return builder.build()
        }
        
        /**
         * 配置开发环境的自签证书
         */
        private fun configureDevelopmentCertificates(
            builder: OkHttpClient.Builder,
            context: Context
        ) {
            try {
                // 创建自定义信任管理器
                val trustManager = CertificateUtils.createCustomTrustManager(context)
                
                // 创建SSL上下文
                val sslContext = SSLContext.getInstance("TLS")
                sslContext.init(null, arrayOf(trustManager), null)
                
                // 配置OkHttp使用自定义SSL
                builder.sslSocketFactory(
                    sslContext.socketFactory,
                    trustManager
                )
                
                // 主机名验证（开发环境可放宽）
                builder.hostnameVerifier { hostname, session ->
                    // 可以添加更严格的主机名验证逻辑
                    true // 临时允许所有主机名（仅限开发环境）
                }
                
            } catch (e: Exception) {
                throw RuntimeException("Failed to configure SSL certificates", e)
            }
        }
        
        /**
         * 获取证书信息用于调试
         */
        fun getCertificateInfo(context: Context): String {
            return try {
                "Certificate Fingerprint: ${CertificateUtils.getCertificateFingerprint(context)}"
            } catch (e: Exception) {
                "Failed to get certificate info: ${e.message}"
            }
        }
    }
}
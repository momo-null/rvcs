package com.rvcs.android

import android.app.Application
import android.util.Log
import okhttp3.OkHttpClient
import java.security.cert.X509Certificate
import java.util.concurrent.TimeUnit
import javax.net.ssl.*

/**
 * RVCS Application
 * 全局初始化 LiveKit 和其他组件
 */
class RVCSApplication : Application() {

    companion object {
        private const val TAG = "RVCSApplication"
    }

    override fun onCreate() {
        super.onCreate()
        Log.d(TAG, "RVCS Application initialized")

        // 初始化 LiveKit（配置信任自签证书）
        initializeLiveKit()
    }

    /**
     * 初始化 LiveKit SDK
     * 配置自定义 SSL 以支持自签证书
     */
    private fun initializeLiveKit() {
        try {
            // 创建信任所有证书的 OkHttpClient
            val okHttpClient = createUnsafeOkHttpClient()
            Log.d(TAG, "Unsafe OkHttpClient created")

            // LiveKit 2.10.0 使用 Room.ConnectOptions 来配置
            // 这里暂时不做配置，由网络安全性配置处理
            Log.d(TAG, "LiveKit will use network security config for SSL")

        } catch (e: Exception) {
            Log.e(TAG, "Failed to initialize LiveKit: ${e.message}", e)
        }
    }

    /**
     * 创建信任所有证书的 OkHttpClient
     * 仅用于开发/测试环境，生产环境应使用有效证书
     */
    private fun createUnsafeOkHttpClient(): OkHttpClient {
        return OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .hostnameVerifier { _, _ -> true }
            .sslSocketFactory(createUnsafeSSLSocketFactory(), createUnsafeTrustManager()[0])
            .build()
    }

    private fun createUnsafeSSLSocketFactory(): SSLSocketFactory {
        val sslContext = SSLContext.getInstance("TLS")
        sslContext.init(null, createUnsafeTrustManager(), java.security.SecureRandom())
        return sslContext.socketFactory
    }

    private fun createUnsafeTrustManager(): Array<X509TrustManager> {
        return arrayOf(object : X509TrustManager {
            override fun checkClientTrusted(chain: Array<X509Certificate>, authType: String) {}

            override fun checkServerTrusted(chain: Array<X509Certificate>, authType: String) {}

            override fun getAcceptedIssuers(): Array<X509Certificate> = emptyArray()
        })
    }
}

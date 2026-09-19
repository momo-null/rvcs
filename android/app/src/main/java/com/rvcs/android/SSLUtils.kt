package com.rvcs.android

import okhttp3.OkHttpClient
import java.security.cert.X509Certificate
import java.util.concurrent.TimeUnit
import javax.net.ssl.*

/**
 * SSL工具类
 * 用于处理自签名证书问题
 * 
 * ⚠️ 安全警告: 此工具信任所有证书，仅适用于开发/测试环境
 * 生产环境必须使用有效的CA证书
 */
object SSLUtils {

    /**
     * 创建信任所有证书的OkHttpClient
     * 警告：仅用于开发/测试环境，生产环境不应使用
     */
    fun createUnsafeOkHttpClient(): OkHttpClient {
        return OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .hostnameVerifier { _, _ -> true } // 忽略主机名验证
            .sslSocketFactory(createUnsafeSSLSocketFactory(), createUnsafeTrustManager()[0])
            .build()
    }

    /**
     * 创建不安全的SSL Socket Factory
     * 信任所有证书
     */
    private fun createUnsafeSSLSocketFactory(): SSLSocketFactory {
        val sslContext = SSLContext.getInstance("TLS")
        sslContext.init(null, createUnsafeTrustManager(), java.security.SecureRandom())
        return sslContext.socketFactory
    }

    /**
     * 创建不安全的TrustManager
     * 信任所有证书
     */
    private fun createUnsafeTrustManager(): Array<X509TrustManager> {
        return arrayOf(object : X509TrustManager {
            override fun checkClientTrusted(chain: Array<X509Certificate>, authType: String) {
                // 不检查客户端证书
            }

            override fun checkServerTrusted(chain: Array<X509Certificate>, authType: String) {
                // 不检查服务器证书
            }

            override fun getAcceptedIssuers(): Array<X509Certificate> {
                return emptyArray()
            }
        })
    }
}

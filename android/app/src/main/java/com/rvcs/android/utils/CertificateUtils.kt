package com.rvcs.android.utils

import android.content.Context
import java.io.InputStream
import java.security.KeyStore
import java.security.cert.CertificateFactory
import java.security.cert.X509Certificate
import javax.net.ssl.TrustManagerFactory
import javax.net.ssl.X509TrustManager

/**
 * 自签证书管理工具类
 * 用于处理HTTPS自签证书的信任配置
 */
class CertificateUtils {
    
    companion object {
        private const val CERTIFICATE_ALIAS = "rvcs_server"
        private const val CERTIFICATE_FILE = "certificates/server.crt"
        
        /**
         * 创建包含自签证书的信任管理器
         */
        fun createCustomTrustManager(context: Context): X509TrustManager {
            try {
                // 从assets加载证书
                val certificateInputStream = context.assets.open(CERTIFICATE_FILE)
                val certificate = loadCertificate(certificateInputStream)
                
                // 创建KeyStore并添加证书
                val keyStore = KeyStore.getInstance(KeyStore.getDefaultType())
                keyStore.load(null, null)
                keyStore.setCertificateEntry(CERTIFICATE_ALIAS, certificate)
                
                // 创建TrustManagerFactory
                val trustManagerFactory = TrustManagerFactory.getInstance(
                    TrustManagerFactory.getDefaultAlgorithm()
                )
                trustManagerFactory.init(keyStore)
                
                // 返回信任管理器
                val trustManagers = trustManagerFactory.trustManagers
                return trustManagers.firstOrNull { it is X509TrustManager } as? X509TrustManager
                    ?: throw IllegalStateException("No X509TrustManager found")
                    
            } catch (e: Exception) {
                throw RuntimeException("Failed to create custom trust manager", e)
            }
        }
        
        /**
         * 从输入流加载X509证书
         */
        private fun loadCertificate(inputStream: InputStream): X509Certificate {
            return inputStream.use { stream ->
                val certificateFactory = CertificateFactory.getInstance("X.509")
                certificateFactory.generateCertificate(stream) as X509Certificate
            }
        }
        
        /**
         * 获取证书指纹（用于验证）
         */
        fun getCertificateFingerprint(context: Context): String {
            try {
                val certificateInputStream = context.assets.open(CERTIFICATE_FILE)
                val certificate = loadCertificate(certificateInputStream)
                val publicKey = certificate.publicKey
                val encoded = publicKey.encoded
                return encoded.joinToString("") { "%02x".format(it) }
            } catch (e: Exception) {
                throw RuntimeException("Failed to get certificate fingerprint", e)
            }
        }
    }
}
package com.rvcs.android.manager

import android.content.Context
import android.util.Log
import com.rvcs.android.utils.CertificateUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.File
import java.net.URL

/**
 * 证书更新管理器
 * 支持从服务器动态下载和更新证书
 */
class CertificateUpdateManager(private val context: Context) {
    
    companion object {
        private const val TAG = "CertificateUpdate"
        private const val CERT_DOWNLOAD_URL = "https://your-server.com/certificates/server.crt"
        private const val LOCAL_CERT_FILE = "updated_server.crt"
    }
    
    /**
     * 检查并更新证书
     */
    suspend fun checkAndUpdateCertificate(): Boolean = withContext(Dispatchers.IO) {
        try {
            Log.d(TAG, "Checking for certificate updates...")
            
            // 下载新证书
            val newCertificate = downloadCertificate()
            if (newCertificate.isNullOrEmpty()) {
                Log.w(TAG, "Failed to download certificate")
                return@withContext false
            }
            
            // 验证证书有效性
            if (!validateCertificate(newCertificate)) {
                Log.w(TAG, "Downloaded certificate is invalid")
                return@withContext false
            }
            
            // 保存证书到本地
            if (saveCertificateToFile(newCertificate)) {
                Log.i(TAG, "Certificate updated successfully")
                return@withContext true
            } else {
                Log.e(TAG, "Failed to save certificate")
                return@withContext false
            }
            
        } catch (e: Exception) {
            Log.e(TAG, "Error updating certificate", e)
            return@withContext false
        }
    }
    
    /**
     * 从服务器下载证书
     */
    private fun downloadCertificate(): String? {
        return try {
            val url = URL(CERT_DOWNLOAD_URL)
            url.openConnection().apply {
                connectTimeout = 5000
                readTimeout = 10000
            }.getInputStream().bufferedReader().use { reader ->
                reader.readText()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to download certificate", e)
            null
        }
    }
    
    /**
     * 验证证书有效性
     */
    private fun validateCertificate(certificate: String): Boolean {
        return try {
            // 基本格式验证
            certificate.contains("-----BEGIN CERTIFICATE-----") &&
            certificate.contains("-----END CERTIFICATE-----")
        } catch (e: Exception) {
            Log.e(TAG, "Certificate validation failed", e)
            false
        }
    }
    
    /**
     * 保存证书到本地文件
     */
    private fun saveCertificateToFile(certificate: String): Boolean {
        return try {
            val file = File(context.filesDir, LOCAL_CERT_FILE)
            file.writeText(certificate)
            Log.d(TAG, "Certificate saved to: ${file.absolutePath}")
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to save certificate to file", e)
            false
        }
    }
    
    /**
     * 获取本地更新的证书文件
     */
    fun getUpdatedCertificateFile(): File? {
        val file = File(context.filesDir, LOCAL_CERT_FILE)
        return if (file.exists()) file else null
    }
    
    /**
     * 清除本地证书缓存
     */
    fun clearCachedCertificate() {
        try {
            val file = File(context.filesDir, LOCAL_CERT_FILE)
            if (file.exists()) {
                file.delete()
                Log.d(TAG, "Cached certificate cleared")
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to clear cached certificate", e)
        }
    }
}
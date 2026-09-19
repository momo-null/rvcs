package com.rvcs.android

import android.content.Context
import android.content.SharedPreferences
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import android.util.Log
import java.security.KeyStore
import java.security.SecureRandom
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.SecretKeySpec

/**
 * 加密偏好设置管理器
 * 使用 AES-GCM 加密敏感数据
 */
class EncryptedPreferences(context: Context, name: String) {

    private val TAG = "EncryptedPreferences"

    private val prefs: SharedPreferences = context.getSharedPreferences(name, Context.MODE_PRIVATE)
    private val keyStore: KeyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply {
        load(null)
    }

    private val secureRandom = SecureRandom()

    companion object {
        private const val ANDROID_KEYSTORE = "AndroidKeyStore"
        private const val AES_KEY_ALIAS = "RVCS_AES_KEY"
        private const val KEY_SIZE = 256
        private const val GCM_IV_LENGTH = 12
        private const val GCM_TAG_LENGTH = 128

        // 键名前缀
        private const val ENCRYPTED_PREFIX = "enc_"
    }

    init {
        // 确保密钥存在
        generateOrGetAESKey()
    }

    /**
     * 生成或获取 AES 密钥
     */
    private fun generateOrGetAESKey(): SecretKey {
        // 尝试从 KeyStore 获取密钥
        keyStore.getKey(AES_KEY_ALIAS, null)?.let {
            Log.d(TAG, "AES key found in KeyStore")
            return it as SecretKey
        }

        // 密钥不存在，生成新密钥
        Log.d(TAG, "Generating new AES key")
        val keyGenerator = KeyGenerator.getInstance(
            KeyProperties.KEY_ALGORITHM_AES,
            ANDROID_KEYSTORE
        )

        val keyGenSpec = KeyGenParameterSpec.Builder(
            AES_KEY_ALIAS,
            KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
        )
            .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
            .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
            .setKeySize(KEY_SIZE)
            .build()

        keyGenerator.init(keyGenSpec)
        return keyGenerator.generateKey()
    }

    /**
     * 加密字符串
     */
    fun encrypt(plaintext: String): String {
        return try {
            val key = generateOrGetAESKey()
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.ENCRYPT_MODE, key)

            // 生成随机 IV
            val iv = ByteArray(GCM_IV_LENGTH)
            secureRandom.nextBytes(iv)

            // 加密
            val ciphertext = cipher.doFinal(plaintext.toByteArray(Charsets.UTF_8))

            // 组合 IV + 密文
            val combined = iv + ciphertext

            // Base64 编码
            Base64.encodeToString(combined, Base64.NO_WRAP)
        } catch (e: Exception) {
            Log.e(TAG, "Encryption failed", e)
            throw SecurityException("Failed to encrypt data", e)
        }
    }

    /**
     * 解密字符串
     */
    fun decrypt(encrypted: String): String {
        return try {
            val key = generateOrGetAESKey()
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")

            // Base64 解码
            val combined = Base64.decode(encrypted, Base64.NO_WRAP)

            // 分离 IV 和密文
            val iv = combined.copyOfRange(0, GCM_IV_LENGTH)
            val ciphertext = combined.copyOfRange(GCM_IV_LENGTH, combined.size)

            // 解密
            val gcmSpec = GCMParameterSpec(GCM_TAG_LENGTH, iv)
            cipher.init(Cipher.DECRYPT_MODE, key, gcmSpec)

            val plaintext = cipher.doFinal(ciphertext)
            String(plaintext, Charsets.UTF_8)
        } catch (e: Exception) {
            Log.e(TAG, "Decryption failed", e)
            throw SecurityException("Failed to decrypt data", e)
        }
    }

    /**
     * 保存加密的字符串
     */
    fun putEncryptedString(key: String, value: String) {
        val encryptedValue = encrypt(value)
        prefs.edit().putString(ENCRYPTED_PREFIX + key, encryptedValue).apply()
        Log.d(TAG, "Encrypted and saved: $key")
    }

    /**
     * 获取并解密字符串
     */
    fun getEncryptedString(key: String, defaultValue: String = ""): String {
        val encryptedValue = prefs.getString(ENCRYPTED_PREFIX + key, null)
        return if (encryptedValue != null) {
            try {
                decrypt(encryptedValue)
            } catch (e: Exception) {
                Log.e(TAG, "Failed to decrypt $key", e)
                defaultValue
            }
        } else {
            defaultValue
        }
    }

    /**
     * 保存普通字符串（不加密）
     */
    fun putString(key: String, value: String) {
        prefs.edit().putString(key, value).apply()
    }

    /**
     * 获取普通字符串
     */
    fun getString(key: String, defaultValue: String = ""): String {
        return prefs.getString(key, defaultValue) ?: defaultValue
    }

    /**
     * 保存整数
     */
    fun putInt(key: String, value: Int) {
        prefs.edit().putInt(key, value).apply()
    }

    /**
     * 获取整数
     */
    fun getInt(key: String, defaultValue: Int = 0): Int {
        return prefs.getInt(key, defaultValue)
    }

    /**
     * 保存布尔值
     */
    fun putBoolean(key: String, value: Boolean) {
        prefs.edit().putBoolean(key, value).apply()
    }

    /**
     * 获取布尔值
     */
    fun getBoolean(key: String, defaultValue: Boolean = false): Boolean {
        return prefs.getBoolean(key, defaultValue)
    }

    /**
     * 删除键
     */
    fun remove(key: String) {
        prefs.edit().remove(ENCRYPTED_PREFIX + key).remove(key).apply()
    }

    /**
     * 清空所有数据
     */
    fun clear() {
        prefs.edit().clear().apply()
        Log.d(TAG, "All preferences cleared")
    }

    /**
     * 检查键是否存在
     */
    fun contains(key: String): Boolean {
        return prefs.contains(ENCRYPTED_PREFIX + key) || prefs.contains(key)
    }
}

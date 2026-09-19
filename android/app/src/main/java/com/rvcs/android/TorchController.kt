package com.rvcs.android

import android.content.Context
import android.hardware.camera2.CameraCharacteristics
import android.hardware.camera2.CameraManager
import android.os.Handler
import android.os.Looper
import android.util.Log
import java.util.concurrent.ConcurrentHashMap

class TorchController(context: Context) {
    private val cameraManager = context.getSystemService(Context.CAMERA_SERVICE) as CameraManager
    private var preferredCameraId: String? = null
    private val torchStateMap = ConcurrentHashMap<String, Boolean>()
    private val callback = object : CameraManager.TorchCallback() {
        override fun onTorchModeChanged(cameraId: String, enabled: Boolean) {
            torchStateMap[cameraId] = enabled
            Log.d(TAG, "Torch state changed: cameraId=$cameraId enabled=$enabled")
        }
    }

    companion object {
        private const val TAG = "TorchController"
    }

    init {
        runCatching {
            cameraManager.registerTorchCallback(callback, Handler(Looper.getMainLooper()))
        }.onFailure {
            Log.w(TAG, "registerTorchCallback failed: ${it.message}")
        }
    }

    fun setEnabled(enabled: Boolean): Boolean {
        val candidates = findTorchCameraIds()
        if (candidates.isEmpty()) {
            Log.w(TAG, "No camera with flashlight capability")
            return false
        }

        val ordered = reorderCandidates(candidates, preferredCameraId)
        if (isAnyCandidateInState(ordered, enabled)) {
            Log.i(TAG, "Torch already in expected state: enabled=$enabled")
            return true
        }

        if (trySetTorch(ordered, enabled)) {
            waitForState(ordered, enabled, 300L)
            return true
        }

        // Some ROMs temporarily reject torch operations right after camera state changes.
        Thread.sleep(120)
        if (trySetTorch(ordered, enabled)) {
            waitForState(ordered, enabled, 300L)
            return true
        }

        // If callback indicates desired state reached, treat as success to avoid false-negative UI.
        if (waitForState(ordered, enabled, 450L)) {
            Log.i(TAG, "Torch reached expected state via callback after set failure: enabled=$enabled")
            return true
        }
        return false
    }

    private fun trySetTorch(cameraIds: List<String>, enabled: Boolean): Boolean {
        var lastError: Throwable? = null
        for (cameraId in cameraIds) {
            val result = runCatching {
                cameraManager.setTorchMode(cameraId, enabled)
                preferredCameraId = cameraId
                Log.i(TAG, "Torch set success: enabled=$enabled cameraId=$cameraId")
                true
            }.getOrElse {
                lastError = it
                Log.w(TAG, "Torch set failed on cameraId=$cameraId enabled=$enabled: ${it.message}")
                false
            }
            if (result) {
                return true
            }
        }
        Log.e(TAG, "Failed to set torch=$enabled after trying ${cameraIds.size} camera(s): ${lastError?.message}")
        return false
    }

    private fun findTorchCameraIds(): List<String> {
        val cameraIds = cameraManager.cameraIdList
        val backFacing = mutableListOf<String>()
        val others = mutableListOf<String>()
        for (cameraId in cameraIds) {
            val chars = cameraManager.getCameraCharacteristics(cameraId)
            val hasFlash = chars.get(CameraCharacteristics.FLASH_INFO_AVAILABLE) == true
            if (!hasFlash) continue
            val facing = chars.get(CameraCharacteristics.LENS_FACING)
            if (facing == CameraCharacteristics.LENS_FACING_BACK) {
                backFacing.add(cameraId)
            } else {
                others.add(cameraId)
            }
        }
        return backFacing + others
    }

    private fun reorderCandidates(candidates: List<String>, preferred: String?): List<String> {
        if (preferred.isNullOrBlank() || !candidates.contains(preferred)) {
            return candidates
        }
        val reordered = mutableListOf(preferred)
        for (id in candidates) {
            if (id != preferred) reordered.add(id)
        }
        return reordered
    }

    private fun isAnyCandidateInState(candidates: List<String>, expected: Boolean): Boolean {
        for (cameraId in candidates) {
            if (torchStateMap[cameraId] == expected) return true
        }
        return false
    }

    private fun waitForState(candidates: List<String>, expected: Boolean, timeoutMs: Long): Boolean {
        val deadline = System.currentTimeMillis() + timeoutMs
        while (System.currentTimeMillis() < deadline) {
            if (isAnyCandidateInState(candidates, expected)) return true
            Thread.sleep(40L)
        }
        return isAnyCandidateInState(candidates, expected)
    }
}

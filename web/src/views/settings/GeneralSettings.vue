<template>
  <div class="general-settings">
    <el-card>
      <template #header>
        <span>System Settings</span>
      </template>

      <el-form label-width="220px" style="max-width: 680px;">
        <el-form-item label="Device Auto Refresh (sec)">
          <el-input-number
            v-model="form.deviceAutoRefreshSec"
            :min="5"
            :max="120"
            :step="1"
          />
          <span class="hint">Used by device list, device detail and heartbeat logs</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSave">Save</el-button>
          <el-button @click="handleReset">Reset</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { getWebSettings, saveWebSettings, resetWebSettings } from '@/utils/settings'

const form = reactive({
  deviceAutoRefreshSec: getWebSettings().deviceAutoRefreshSec
})

const handleSave = () => {
  const saved = saveWebSettings({
    deviceAutoRefreshSec: form.deviceAutoRefreshSec
  })
  form.deviceAutoRefreshSec = saved.deviceAutoRefreshSec
  ElMessage.success('Saved')
}

const handleReset = () => {
  const defaults = resetWebSettings()
  form.deviceAutoRefreshSec = defaults.deviceAutoRefreshSec
  ElMessage.success('Reset to default')
}
</script>

<style scoped lang="scss">
.general-settings {
  padding: 20px;

  @media (max-width: 768px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }
}

.hint {
  margin-left: 12px;
  color: #666;
  font-size: 12px;
}

:deep(.el-card) {
  @media (max-width: 768px) {
    margin-bottom: 12px !important;

    .el-card__header {
      padding: 15px;
    }

    .el-card__body {
      padding: 15px;
    }
  }

  @media (max-width: 480px) {
    margin-bottom: 8px !important;

    .el-card__header {
      padding: 12px;
    }

    .el-card__body {
      padding: 12px;
    }
  }
}
</style>


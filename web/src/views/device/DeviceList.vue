<template>
  <div class="device-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>设备列表</span>
          <el-button type="primary" @click="handleAdd">添加设备</el-button>
        </div>
      </template>
      <el-table :data="devices" v-loading="loading" stripe>
        <el-table-column prop="id" label="设备ID" width="200" show-overflow-tooltip />
        <el-table-column prop="name" label="设备名称" min-width="120" />
        <el-table-column prop="ip_address" label="IP地址" width="140" />
        <el-table-column prop="manufacturer" label="制造商" width="120" show-overflow-tooltip />
        <el-table-column prop="android_version" label="Android版本" width="120" />
        <el-table-column prop="app_version" label="应用版本" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="last_seen" label="最后在线" width="180">
          <template #default="{ row }">
            {{ formatTime(row.last_seen) }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadDevices"
        @current-change="loadDevices"
        style="margin-top: 20px"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deviceApi } from '@/api/device'
import type { Device } from '@/api/device'
import dayjs from 'dayjs'
import { getDeviceAutoRefreshIntervalMs } from '@/utils/settings'

const router = useRouter()
const loading = ref(false)
const devices = ref<Device[]>([])
let autoRefreshTimer: number | null = null
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss')

const getStatusType = (status: string) => {
  const types: Record<string, any> = {
    online: 'success',
    offline: 'info',
    error: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    online: '在线',
    offline: '离线',
    error: '错误'
  }
  return texts[status] || '未知'
}

const loadDevices = async () => {
  try {
    loading.value = true
    const result = await deviceApi.getDeviceList({
      page: pagination.page,
      page_size: pagination.page_size
    })
    devices.value = result.items || []
    pagination.total = result.total
  } catch {
    ElMessage.error('加载设备列表失败')
  } finally {
    loading.value = false
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()
  autoRefreshTimer = window.setInterval(() => {
    if (document.visibilityState !== 'visible') return
    loadDevices()
  }, getDeviceAutoRefreshIntervalMs())
}

const stopAutoRefresh = () => {
  if (autoRefreshTimer !== null) {
    window.clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
}

const handleVisibilityChange = () => {
  if (document.visibilityState === 'visible') loadDevices()
}

const handleAdd = () => {
  ElMessage.info('添加设备功能开发中')
}

const handleView = (row: Device) => router.push(`/devices/${row.id}`)

const handleEdit = (_row: Device) => {
  ElMessage.info('编辑设备功能开发中')
}

const handleDelete = async (row: Device) => {
  try {
    await ElMessageBox.confirm('确定要删除该设备吗？', '提示', {
      type: 'warning'
    })
    await deviceApi.deleteDevice(row.id)
    ElMessage.success('删除成功')
    loadDevices()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadDevices()
  startAutoRefresh()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  stopAutoRefresh()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped lang="scss">
.device-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  @media (max-width: 768px) {
    .card-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .el-button {
        width: 100%;
      }
    }

    :deep(.el-table) {
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
    .card-header {
      gap: 8px;
      font-size: 14px;

      .el-button {
        font-size: 12px;
      }
    }

    :deep(.el-table) {
      font-size: 12px;

      .el-table__cell {
        padding: 8px 0;
      }
    }

    :deep(.el-pagination) {
      font-size: 12px;

      .el-pager li {
        min-width: 28px;
        height: 28px;
        line-height: 28px;
      }
    }
  }
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

<template>
  <div class="alert-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>告警列表</span>
          <el-select v-model="statusFilter" placeholder="状态筛选" style="width: 150px" @change="loadAlerts">
            <el-option label="全部" value="" />
            <el-option label="待处理" value="pending" />
            <el-option label="已确认" value="acknowledged" />
          </el-select>
        </div>
      </template>
      <el-table :data="alerts" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="device_id" label="设备ID" />
        <el-table-column prop="event_type" label="类型" width="160" />
        <el-table-column prop="message" label="消息" />
        <el-table-column prop="level" label="级别" width="100">
          <template #default="{ row }">
            <el-tag :type="getAlertLevelType(row.level)">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getAlertStatusType(row.acknowledged)">
              {{ row.acknowledged ? 'acknowledged' : 'pending' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button v-if="!row.acknowledged" link type="primary" @click="handleAcknowledge(row)">确认</el-button>
            <el-button v-if="!row.acknowledged" link type="success" @click="handleResolve(row)">解决</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadAlerts"
        @current-change="loadAlerts"
        style="margin-top: 20px"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { alertApi } from '@/api/alert'
import type { Alert } from '@/api/alert'
import dayjs from 'dayjs'

const loading = ref(false)
const alerts = ref<Alert[]>([])
const statusFilter = ref('')
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const getAlertLevelType = (level: string) => {
  const types: Record<string, any> = {
    info: 'info',
    warning: 'warning',
    error: 'danger',
    critical: 'danger'
  }
  return types[level] || 'info'
}

const getAlertStatusType = (acknowledged: boolean) => {
  return acknowledged ? 'info' : 'warning'
}

const loadAlerts = async () => {
  try {
    loading.value = true
    const acknowledged = statusFilter.value === '' ? undefined : statusFilter.value === 'acknowledged'
    const result = await alertApi.getAlerts({
      page: pagination.page,
      page_size: pagination.page_size,
      acknowledged
    })
    alerts.value = result.alerts
    pagination.total = result.total
  } catch (error) {
    ElMessage.error('加载告警列表失败')
  } finally {
    loading.value = false
  }
}

const handleAcknowledge = async (row: Alert) => {
  try {
    await alertApi.acknowledgeAlert(row.id)
    ElMessage.success('确认成功')
    loadAlerts()
  } catch (error) {
    ElMessage.error('确认失败')
  }
}

const handleResolve = async (row: Alert) => {
  try {
    await alertApi.resolveAlert(row.id)
    ElMessage.success('解决成功')
    loadAlerts()
  } catch (error) {
    ElMessage.error('解决失败')
  }
}

onMounted(() => {
  loadAlerts()
})
</script>

<style scoped lang="scss">
.alert-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    // 响应式：移动端换行显示
    @media (max-width: 768px) {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .el-select {
        width: 100%;
      }
    }

    @media (max-width: 480px) {
      gap: 8px;
      font-size: 14px;

      .el-button {
        font-size: 12px;
      }
    }
  }

  // 响应式：表格适配
  @media (max-width: 768px) {
    :deep(.el-table) {
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
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

// 响应式：el-card 在移动端适配
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

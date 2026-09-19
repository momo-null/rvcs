<template>
  <div class="activity-log">
    <el-card>
      <template #header>
        <span>活动日志</span>
      </template>
      <el-table :data="logs" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="device_name" label="设备名称" />
        <el-table-column prop="action" label="操作" />
        <el-table-column prop="details" label="详情" />
        <el-table-column prop="user" label="用户" width="120" />
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadLogs"
        @current-change="loadLogs"
        style="margin-top: 20px"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { getActivityLogs } from '@/api/log'
import type { ActivityLog } from '@/api/log'
import dayjs from 'dayjs'

const loading = ref(false)
const logs = ref<ActivityLog[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const loadLogs = async () => {
  try {
    loading.value = true
    const result = await getActivityLogs({
      page: pagination.page,
      page_size: pagination.page_size
    })
    logs.value = result.items
    pagination.total = result.total
  } catch (error) {
    ElMessage.error('加载活动日志失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadLogs()
})
</script>

<style scoped lang="scss">
.activity-log {
  padding: 20px;

  // 响应式：移动端减�?padding
  @media (max-width: 768px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }
}

// 响应式：表格适配
:deep(.el-table) {
  @media (max-width: 768px) {
    font-size: 13px;
  }

  @media (max-width: 480px) {
    font-size: 12px;

    .el-table__cell {
      padding: 8px 0;
    }
  }

  @media (max-width: 768px) {
    :deep(.el-pagination) {
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
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


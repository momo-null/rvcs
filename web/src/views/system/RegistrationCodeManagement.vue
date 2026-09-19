<template>
  <div class="registration-code-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>注册码管理</span>
          <div>
            <el-button type="warning" @click="handleCleanup" :loading="cleaning">清理过期</el-button>
            <el-button type="primary" @click="handleGenerate">生成注册码</el-button>
          </div>
        </div>
      </template>
      <el-table :data="codes" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="注册码" width="200">
          <template #default="{ row }">
            <el-tag>{{ row.code }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="备注" />
        <el-table-column prop="created_by" label="创建者" width="120" />
        <el-table-column prop="expires_at" label="过期时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.expires_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending' || row.status === 'used' || row.status === 'revoked'"
              link
              type="primary"
              @click="handleReset(row)"
            >
              重置
            </el-button>
            <el-button
              v-if="row.status === 'pending'"
              link
              type="warning"
              @click="handleRevoke(row)"
            >
              撤销
            </el-button>
            <el-button
              link
              type="danger"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadCodes"
        @current-change="loadCodes"
        style="margin-top: 20px"
      />
    </el-card>

    <!-- 生成注册码对话框 -->
    <el-dialog v-model="dialogVisible" title="生成注册码" width="500px">
      <el-form :model="form" label-width="120px">
        <el-form-item label="有效天数">
          <el-input-number v-model="form.validity_days" :min="1" :max="365" />
          <span style="margin-left: 10px; color: #999;">默认30天</span>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="可选，用于说明注册码用途" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmGenerate" :loading="generating">
          生成
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置注册码对话框 -->
    <el-dialog v-model="resetDialogVisible" title="重置注册码" width="500px">
      <el-form :model="resetForm" label-width="120px">
        <el-form-item label="延长过期时间">
          <el-checkbox v-model="resetForm.extendTime" />
        </el-form-item>
        <el-form-item v-if="resetForm.extendTime" label="延长天数">
          <el-input-number v-model="resetForm.extendDays" :min="1" :max="365" />
        </el-form-item>
      </el-form>
      <el-alert
        title="重置后注册码将变为活跃状态，设备关联将被清空"
        type="warning"
        :closable="false"
        style="margin-top: 15px"
      />
      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmReset" :loading="resetting">
          确认重置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { registrationCodeApi } from '@/api/registrationCode'
import type { RegistrationCode } from '@/api/registrationCode'
import dayjs from 'dayjs'

const loading = ref(false)
const codes = ref<RegistrationCode[]>([])
const dialogVisible = ref(false)
const resetDialogVisible = ref(false)
const generating = ref(false)
const resetting = ref(false)
const cleaning = ref(false)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const form = reactive({
  validity_days: 30,
  description: ''
})

const resetForm = reactive({
  extendTime: false,
  extendDays: 30
})

const currentCodeId = ref<string>('')

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const getStatusType = (status: string) => {
  const types: Record<string, any> = {
    pending: 'success',
    expired: 'info',
    used: 'warning',
    revoked: 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    pending: '活跃',
    expired: '已过期',
    used: '已使用',
    revoked: '已撤销'
  }
  return texts[status] || status
}

const loadCodes = async () => {
  try {
    loading.value = true
    const result = await registrationCodeApi.getCodes({
      page: pagination.page,
      page_size: pagination.page_size
    })
    codes.value = result.codes
    pagination.total = result.total
  } catch (error) {
    ElMessage.error('加载注册码列表失败')
  } finally {
    loading.value = false
  }
}

const handleGenerate = () => {
  form.validity_days = 30
  form.description = ''
  dialogVisible.value = true
}

const handleConfirmGenerate = async () => {
  try {
    generating.value = true
    await registrationCodeApi.generateCode({
      validity_days: form.validity_days,
      description: form.description || undefined
    })
    ElMessage.success('注册码生成成功')
    dialogVisible.value = false
    loadCodes()
  } catch (error) {
    ElMessage.error('生成注册码失败')
  } finally {
    generating.value = false
  }
}

const handleRevoke = async (row: RegistrationCode) => {
  try {
    await ElMessageBox.confirm('确定要撤销该注册码吗？', '提示', {
      type: 'warning'
    })
    await registrationCodeApi.revokeCode(row.id)
    ElMessage.success('撤销成功')
    loadCodes()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('撤销失败')
    }
  }
}

const handleReset = (row: RegistrationCode) => {
  currentCodeId.value = row.id
  resetForm.extendTime = false
  resetForm.extendDays = 30
  resetDialogVisible.value = true
}

const handleConfirmReset = async () => {
  try {
    resetting.value = true
    const data = resetForm.extendTime ? { extend_days: resetForm.extendDays } : {}
    await registrationCodeApi.resetCode(currentCodeId.value, data)
    ElMessage.success('重置成功')
    resetDialogVisible.value = false
    loadCodes()
  } catch (error: any) {
    ElMessage.error('重置失败')
  } finally {
    resetting.value = false
  }
}

const handleDelete = async (row: RegistrationCode) => {
  try {
    await ElMessageBox.confirm('确定要删除该注册码吗？此操作不可恢复。', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    await registrationCodeApi.deleteCode(row.id)
    ElMessage.success('删除成功')
    loadCodes()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleCleanup = async () => {
  try {
    await ElMessageBox.confirm('确定要清理所有过期的注册码吗？此操作不可恢复。', '提示', {
      type: 'warning'
    })
    cleaning.value = true
    const result = await registrationCodeApi.cleanupExpiredCodes()
    ElMessage.success(`成功清理 ${result.deleted_count} 个过期注册码`)
    loadCodes()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('清理失败')
    }
  } finally {
    cleaning.value = false
  }
}

onMounted(() => {
  loadCodes()
})
</script>

<style scoped lang="scss">
.registration-code-management {
  padding: 20px;

  // 响应式：移动端减少 padding
  @media (max-width: 768px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    // 响应式：移动端换行显示
    @media (max-width: 768px) {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      > div {
        display: flex;
        gap: 10px;
        width: 100%;

        .el-button {
          flex: 1;
        }
      }
    }

    @media (max-width: 480px) {
      gap: 8px;
      font-size: 14px;

      > div {
        gap: 8px;
        flex-direction: column;

        .el-button {
          width: 100%;
          font-size: 12px;
        }
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

// 响应式：el-dialog 在移动端适配
:deep(.el-dialog) {
  @media (max-width: 768px) {
    width: 90% !important;
    margin: 5vh auto;

    .el-dialog__header {
      padding: 15px;
    }

    .el-dialog__body {
      padding: 15px;
    }

    .el-dialog__footer {
      padding: 15px;
    }

    .el-form-item__label {
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
    width: 95% !important;

    .el-dialog__header {
      padding: 12px;
      font-size: 14px;
    }

    .el-dialog__body {
      padding: 12px;
    }

    .el-dialog__footer {
      padding: 12px;

      .el-button {
        font-size: 12px;
      }
    }

    .el-form-item {
      margin-bottom: 16px;

      .el-form-item__label {
        font-size: 12px;
        margin-bottom: 6px;
      }
    }
  }
}
</style>

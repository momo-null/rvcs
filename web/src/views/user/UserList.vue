<template>
  <div class="user-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户列表</span>
          <el-button type="primary" @click="handleAdd">添加用户</el-button>
        </div>
      </template>
      <el-table :data="users" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" />
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="role" label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="row.role === 'admin' ? 'danger' : 'primary'">
              {{ row.role === 'admin' ? '管理员' : '用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="warning" @click="handleResetPassword(row)">修改密码</el-button>
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
        @size-change="loadUsers"
        @current-change="loadUsers"
        style="margin-top: 20px"
      />
    </el-card>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="500px">
      <el-form :model="passwordForm" label-width="100px" :rules="passwordRules" ref="passwordFormRef">
        <el-form-item label="用户名">
          <el-input v-model="selectedUser.username" disabled />
        </el-form-item>
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="passwordForm.new_password" type="password" show-password placeholder="请输入新密码（至少6位）" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirm_password">
          <el-input v-model="passwordForm.confirm_password" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmResetPassword" :loading="resetting">
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { userApi } from '@/api/user'
import type { User } from '@/api/user'
import dayjs from 'dayjs'

const loading = ref(false)
const users = ref<User[]>([])
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

const passwordDialogVisible = ref(false)
const resetting = ref(false)
const passwordFormRef = ref<FormInstance>()
const selectedUser = ref<Partial<User>>({})
const passwordForm = reactive({
  new_password: '',
  confirm_password: ''
})

const validateConfirmPassword = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== passwordForm.new_password) {
    callback(new Error('两次输入密码不一致'))
  } else {
    callback()
  }
}

const passwordRules: FormRules = {
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const loadUsers = async () => {
  try {
    loading.value = true
    const result = await userApi.getUserList({
      page: pagination.page,
      page_size: pagination.page_size
    })
    users.value = result.users
    pagination.total = result.total
  } catch (error) {
    ElMessage.error('加载用户列表失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  ElMessage.info('添加用户功能开发中')
}

const handleEdit = (row: User) => {
  ElMessage.info('编辑用户功能开发中')
}

const handleResetPassword = (row: User) => {
  selectedUser.value = row
  passwordForm.new_password = ''
  passwordForm.confirm_password = ''
  passwordDialogVisible.value = true
}

const handleConfirmResetPassword = async () => {
  if (!passwordFormRef.value) return

  try {
    await passwordFormRef.value.validate()
    resetting.value = true
    await userApi.resetUserPassword(selectedUser.value.id!, { new_password: passwordForm.new_password })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
    passwordFormRef.value.resetFields()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('密码修改失败')
    }
  } finally {
    resetting.value = false
  }
}

const handleDelete = async (row: User) => {
  try {
    await ElMessageBox.confirm('确定要删除该用户吗？', '提示', {
      type: 'warning'
    })
    await userApi.deleteUser(row.id)
    ElMessage.success('删除成功')
    loadUsers()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped lang="scss">
.user-list {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  // 响应式：移动端优化
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

    :deep(.el-dialog) {
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

    :deep(.el-dialog) {
      width: 95% !important;

      .el-form-item__label {
        font-size: 12px;
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

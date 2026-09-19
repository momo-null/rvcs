<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <div class="card-header">
          <h2>RVCS - Remote Vision & Control System</h2>
          <p>用户注册</p>
        </div>
      </template>
      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="registerRules"
        label-position="top"
        @submit.prevent="handleRegister"
      >
        <el-form-item label="用户名" prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="请输入用户名"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="registerForm.email"
            placeholder="请输入邮箱（可选）"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><Message /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="registerForm.password"
            type="password"
            placeholder="请输入密码"
            size="large"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
          <!-- 密码强度指示器 -->
          <PasswordStrength :password="registerForm.password" />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            size="large"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="注册码（可选）" prop="registration_code">
          <el-input
            v-model="registerForm.registration_code"
            placeholder="请输入注册码（可选）"
            size="large"
            clearable
          >
            <template #prefix>
              <el-icon><Key /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>
      </el-form>
      <div class="login-link">
        已有账号？<router-link to="/login">立即登录</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock, Message, Key } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import PasswordStrength from '@/components/PasswordStrength.vue'

const router = useRouter()
const authStore = useAuthStore()

const registerFormRef = ref<FormInstance>()
const loading = ref(false)

const registerForm = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  registration_code: ''
})

const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== registerForm.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在3到20个字符', trigger: 'blur' }
  ],
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleRegister = async () => {
  if (!registerFormRef.value) return

  try {
    await registerFormRef.value.validate()
    loading.value = true

    const { confirmPassword, ...registerData } = registerForm
    await authStore.register(registerData)

    ElMessage.success('注册成功')
    router.push('/')
  } catch (error: any) {
    if (error?.response?.data?.message) {
      ElMessage.error(error.response.data.message)
    } else {
      ElMessage.error('注册失败，请重试')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  // 响应式：移动端调整 padding
  @media (max-width: 768px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }
}

.register-card {
  width: 450px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);

  // 响应式：移动端调整卡片宽度
  @media (max-width: 768px) {
    width: 100%;
    max-width: 500px;
  }

  @media (max-width: 480px) {
    width: 100%;
    max-width: 400px;
  }
}

.card-header {
  text-align: center;

  h2 {
    margin: 0 0 5px;
    color: #303133;

    // 响应式：移动端调整标题字体
    @media (max-width: 768px) {
      font-size: 20px;
    }

    @media (max-width: 480px) {
      font-size: 18px;
    }
  }

  p {
    margin: 0;
    font-size: 14px;
    color: #909399;

    // 响应式：移动端调整副标题字体
    @media (max-width: 480px) {
      font-size: 12px;
    }
  }
}

.login-link {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;

  // 响应式：移动端调整字体
  @media (max-width: 480px) {
    margin-top: 15px;
    font-size: 12px;
  }

  a {
    color: #409eff;
    text-decoration: none;

    &:hover {
      text-decoration: underline;
    }
  }
}

// 响应式：el-card 在移动端适配
:deep(.el-card) {
  @media (max-width: 768px) {
    .el-card__header {
      padding: 20px;
    }

    .el-card__body {
      padding: 20px;
    }
  }

  @media (max-width: 480px) {
    .el-card__header {
      padding: 15px;
    }

    .el-card__body {
      padding: 15px;
    }

    .el-form-item {
      margin-bottom: 16px;

      .el-form-item__label {
        font-size: 13px;
        margin-bottom: 6px;
      }
    }

    .el-button--large {
      font-size: 14px;
      padding: 12px 20px;
    }
  }
}
</style>

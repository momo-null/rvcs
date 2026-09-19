<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <div class="card-header">
          <h2>RVCS - Remote Vision & Control System</h2>
          <p>Sign in to continue</p>
        </div>
      </template>

      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        label-position="top"
        @submit.prevent="handleLogin"
      >
        <el-form-item label="Username" prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="Enter username"
            size="large"
            clearable
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="Password" prop="password">
          <el-input
            v-model="loginForm.password"
            type="password"
            placeholder="Enter password"
            size="large"
            show-password
            @keyup.enter="handleLogin"
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item class="remember-row">
          <el-checkbox v-model="rememberPassword">Remember password</el-checkbox>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            @click="handleLogin"
          >
            Sign in
          </el-button>
        </el-form-item>
      </el-form>

      <div class="register-link">
        No account yet? <router-link to="/register">Register now</router-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const REMEMBER_PASSWORD_KEY = 'remember_password'
const REMEMBER_USERNAME_KEY = 'remember_username'
const REMEMBER_PLAIN_PASSWORD_KEY = 'remember_plain_password'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref<FormInstance>()
const loading = ref(false)
const rememberPassword = ref(false)

const loginForm = reactive({
  username: '',
  password: ''
})

const loginRules: FormRules = {
  username: [{ required: true, message: 'Please enter username', trigger: 'blur' }],
  password: [
    { required: true, message: 'Please enter password', trigger: 'blur' },
    { min: 6, message: 'Password must be at least 6 characters', trigger: 'blur' }
  ]
}

const restoreRememberedCredentials = () => {
  const savedEnabled = localStorage.getItem(REMEMBER_PASSWORD_KEY) === '1'
  if (!savedEnabled) return

  const savedUsername = localStorage.getItem(REMEMBER_USERNAME_KEY) || ''
  const savedPassword = localStorage.getItem(REMEMBER_PLAIN_PASSWORD_KEY) || ''

  rememberPassword.value = true
  loginForm.username = savedUsername
  loginForm.password = savedPassword
}

const persistRememberedCredentials = () => {
  if (rememberPassword.value) {
    localStorage.setItem(REMEMBER_PASSWORD_KEY, '1')
    localStorage.setItem(REMEMBER_USERNAME_KEY, loginForm.username)
    localStorage.setItem(REMEMBER_PLAIN_PASSWORD_KEY, loginForm.password)
  } else {
    localStorage.removeItem(REMEMBER_PASSWORD_KEY)
    localStorage.removeItem(REMEMBER_USERNAME_KEY)
    localStorage.removeItem(REMEMBER_PLAIN_PASSWORD_KEY)
  }
}

const handleLogin = async () => {
  if (!loginFormRef.value) return

  try {
    await loginFormRef.value.validate()
    loading.value = true

    persistRememberedCredentials()
    await authStore.login(loginForm)

    ElMessage.success('Login successful')
    router.push('/')
  } catch (error: any) {
    if (error?.response?.data?.message) {
      ElMessage.error(error.response.data.message)
    } else {
      ElMessage.error('Login failed, please check username and password')
    }
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  restoreRememberedCredentials()
})
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;

  @media (max-width: 768px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }
}

.login-card {
  width: 400px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);

  @media (max-width: 768px) {
    width: 100%;
    max-width: 450px;
  }

  @media (max-width: 480px) {
    width: 100%;
    max-width: 380px;
  }
}

.card-header {
  text-align: center;

  h2 {
    margin: 0 0 5px;
    color: #303133;

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

    @media (max-width: 480px) {
      font-size: 12px;
    }
  }
}

.remember-row {
  margin-top: -8px;
}

.register-link {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;

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
      margin-bottom: 18px;

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

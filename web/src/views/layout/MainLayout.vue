<template>
  <el-container class="main-layout">
    <!-- 绉诲姩绔伄缃╁眰 -->
    <div v-if="isMobile && sidebarVisible" class="sidebar-overlay" @click="toggleSidebar"></div>

    <!-- 渚ц竟鏍?-->
    <el-aside :class="['sidebar', { 'mobile-visible': isMobile && sidebarVisible }]">
      <div class="logo">
        <h2>RVCS</h2>
        <p>Remote Vision & Control System</p>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#001529"
        text-color="#fff"
        active-text-color="#1890ff"
      >
        <el-menu-item index="/">
          <el-icon><component :is="icons['HomeFilled']" /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/devices">
          <el-icon><component :is="icons['Monitor']" /></el-icon>
          <span>设备管理</span>
        </el-menu-item>
        <el-menu-item index="/alerts">
          <el-icon><component :is="icons['Warning']" /></el-icon>
          <span>告警管理</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/users">
          <el-icon><component :is="icons['User']" /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/registration-codes">
          <el-icon><component :is="icons['Key']" /></el-icon>
          <span>注册码管理</span>
        </el-menu-item>
        <el-menu-item index="/logs">
          <el-icon><component :is="icons['Document']" /></el-icon>
          <span>活动日志</span>
        </el-menu-item>
        <el-menu-item v-if="isAdmin" index="/settings">
          <el-icon><component :is="icons['Setting']" /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header>
        <div class="header-content">
          <div class="header-left">
            <!-- 姹夊牎鑿滃崟鎸夐挳 -->
            <el-button
              v-if="isMobile"
              class="hamburger-btn"
              text
              @click="toggleSidebar"
            >
              <el-icon size="24"><component :is="icons['Menu']" /></el-icon>
            </el-button>
            <div class="breadcrumb">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item>{{ currentPageName }}</el-breadcrumb-item>
              </el-breadcrumb>
            </div>
          </div>
          <div class="user-info">
            <el-dropdown>
              <span class="user-name">
                <el-icon><component :is="icons['User']" /></el-icon>
                {{ user?.username }}
              </span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="logout">
                    <el-icon><component :is="icons['SwitchButton']" /></el-icon>
                    退出登录
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

const icons = ElementPlusIconsVue
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

// 鍝嶅簲寮忎晶杈规爮鐘舵€?
const isMobile = ref(false)
const sidebarVisible = ref(false)

const activeMenu = computed(() => route.path)
const user = computed(() => authStore.user)
const isAdmin = computed(() => authStore.isAdmin)

const currentPageName = computed(() => {
  const nameMap: Record<string, string> = {
    '/': '仪表盘',
    '/devices': '设备管理',
    '/alerts': '告警管理',
    '/users': '用户管理',
    '/registration-codes': '注册码管理',
    '/logs': '活动日志',
    '/settings': '系统设置'
  }
  return nameMap[route.path] || '仪表盘'
})

// 妫€鏌ユ槸鍚︽槸绉诲姩璁惧
const isMobileDevice = () => {
  // 妫€娴嬭Е鎽歌兘鍔?
  const isTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0

  // 妫€娴?User Agent
  const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera
  const isMobileUA = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(userAgent)
  const isTabletUA = /iPad|Android(?!.*Mobile)|Tablet/i.test(userAgent)

  // 妫€娴嬪睆骞曞搴︼紙杈呭姪鍒ゆ柇锛?
  const screenWidth = window.innerWidth
  const isSmallScreen = screenWidth < 1024

  // 缁煎悎鍒ゆ柇锛氳Е鎽稿睆 + (绉诲姩璁惧UA 鎴?灏忓睆骞?
  return isTouch && (isMobileUA || isTabletUA || isSmallScreen)
}

// 妫€鏌ヨ澶囩被鍨?
const checkMobile = () => {
  const mobile = isMobileDevice()
  console.log('Info log', {
    userAgent: navigator.userAgent,
    touch: 'ontouchstart' in window,
    touchPoints: navigator.maxTouchPoints,
    screenWidth: window.innerWidth,
    isMobile: mobile
  })
  isMobile.value = mobile
  // 妗岄潰绔粯璁ゆ樉绀轰晶杈规爮锛岀Щ鍔ㄧ榛樿闅愯棌
  if (!mobile) {
    sidebarVisible.value = true
  } else {
    sidebarVisible.value = false
  }
}

// 鍒囨崲渚ц竟鏍?
const toggleSidebar = () => {
  sidebarVisible.value = !sidebarVisible.value
}

// 鐩戝惉璺敱鍙樺寲锛岀Щ鍔ㄧ鏃跺叧闂晶杈规爮
watch(() => route.path, () => {
  if (isMobile.value) {
    sidebarVisible.value = false
  }
})

const logout = () => {
  authStore.logout()
  router.push('/login')
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
</script>

<style scoped lang="scss">
.main-layout {
  height: 100vh;
  position: relative;
}

// 绉诲姩绔伄缃╁眰
.sidebar-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 999;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

// 渚ц竟鏍?
:deep(.el-aside.sidebar) {
  background-color: #001529;
  color: #fff;
  overflow-x: hidden;
  width: 250px !important;
  height: 100%;
  transition: transform 0.3s ease, width 0.3s ease;
  z-index: 1000;

  .logo {
    padding: 20px;
    text-align: center;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    h2 {
      margin: 0;
      font-size: 24px;
      font-weight: bold;
    }

    p {
      margin: 5px 0 0;
      font-size: 12px;
      opacity: 0.8;
    }
  }

  .el-menu {
    border: none;
  }

  // 鍝嶅簲寮忥細骞虫澘鍜屾墜鏈虹渚ц竟鏍忛殣钘?
  @media (max-width: 1024px) {
    width: 220px !important;
    position: fixed !important;
    left: 0;
    top: 0;
    transform: translateX(-100%) !important;

    &.mobile-visible {
      transform: translateX(0) !important;
    }

    .logo {
      padding: 15px;

      h2 {
        font-size: 20px;
      }

      p {
        font-size: 11px;
      }
    }

    .el-menu-item {
      font-size: 13px;

      .el-icon {
        font-size: 16px;
      }
    }
  }

  @media (max-width: 480px) {
    width: 240px !important;

    .logo {
      padding: 12px;

      h2 {
        font-size: 18px;
      }

      p {
        font-size: 10px;
      }
    }

    .el-menu-item {
      font-size: 12px;
      padding: 0 15px !important;

      .el-icon {
        font-size: 16px;
      }

      span {
        margin-left: 10px;
      }
    }
  }
}

.el-header {
  background-color: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  padding: 0 20px;
  height: 60px;

  .header-content {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .hamburger-btn {
    padding: 8px;
    margin-right: 0;

    &:hover {
      background-color: #f5f5f5;
    }
  }

  .user-name {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;

    .el-icon {
      font-size: 20px;
    }
  }

  // 鍝嶅簲寮忥細绉诲姩绔ご閮ㄩ珮搴﹀拰闂磋窛璋冩暣
  @media (max-width: 1024px) {
    height: 50px;
    padding: 0 15px;

    .header-content {
      .breadcrumb {
        font-size: 13px;
      }

      .user-name {
        font-size: 13px;

        .el-icon {
          font-size: 18px;
        }
      }
    }
  }

  @media (max-width: 480px) {
    height: 45px;
    padding: 0 10px;

    .header-content {
      .header-left {
        gap: 8px;
      }

      .breadcrumb {
        font-size: 12px;
      }

      .user-name {
        font-size: 12px;
        gap: 6px;

        .el-icon {
          font-size: 16px;
        }
      }
    }
  }
}

.el-main {
  background-color: #f0f2f5;
  padding: 20px;

  // 鍝嶅簲寮忥細绉诲姩绔噺灏?padding
  @media (max-width: 1024px) {
    padding: 15px;
  }

  @media (max-width: 480px) {
    padding: 10px;
  }
}
</style>



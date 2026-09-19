import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/auth/Login.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/register',
      redirect: '/login'
    },
    {
      path: '/',
      component: () => import('@/views/layout/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/Dashboard.vue')
        },
        {
          path: 'devices',
          name: 'DeviceList',
          component: () => import('@/views/device/DeviceList.vue')
        },
        {
          path: 'devices/:id',
          name: 'DeviceDetail',
          component: () => import('@/views/device/DeviceDetail.vue')
        },
        {
          path: 'alerts',
          name: 'AlertList',
          component: () => import('@/views/alert/AlertList.vue')
        },
        {
          path: 'users',
          name: 'UserList',
          component: () => import('@/views/user/UserList.vue'),
          meta: { requiresAdmin: true }
        },
        {
          path: 'logs',
          name: 'ActivityLog',
          component: () => import('@/views/log/ActivityLog.vue')
        },
        {
          path: 'registration-codes',
          name: 'RegistrationCodeManagement',
          component: () => import('@/views/system/RegistrationCodeManagement.vue'),
          meta: { requiresAdmin: true }
        },
        {
          path: 'settings',
          name: 'GeneralSettings',
          component: () => import('@/views/settings/GeneralSettings.vue')
        }
      ]
    }
  ]
})

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // Load auth state from storage
  authStore.loadFromStorage()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
  } else if (to.meta.requiresAdmin && !authStore.isAdmin) {
    next('/')
  } else {
    next()
  }
})

export default router


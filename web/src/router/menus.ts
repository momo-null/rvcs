import type { RouteRecordRaw } from 'vue-router'

interface MenuItem {
  title: string
  icon?: string
  path?: string
  children?: MenuItem[]
  roles?: string[]
}

export const menuRoutes: MenuItem[] = [
  {
    title: '系统管理',
    icon: 'Setting',
    children: [
      {
        title: '用户管理',
        path: '/users',
        icon: 'User',
        roles: ['admin']
      },
      {
        title: '注册码管理',
        path: '/registration-codes',
        icon: 'Key',
        roles: ['admin']
      },
      {
        title: '系统设置',
        path: '/settings',
        icon: 'Tools',
        roles: ['admin']
      }
    ]
  }
]

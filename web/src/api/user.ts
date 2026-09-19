import request from './request'

export interface User {
  id: number
  username: string
  email: string
  role: string
  created_at: string
  updated_at: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface ResetPasswordRequest {
  new_password: string
}

export const userApi = {
  // 获取用户列表
  getUserList(params?: { page?: number; page_size?: number }): Promise<{
    users: User[]
    total: number
  }> {
    return request.get('/users', { params })
  },

  // 获取用户详情
  getUser(id: number): Promise<User> {
    return request.get(`/users/${id}`)
  },

  // 创建用户
  createUser(data: { username: string; password: string; email?: string; role?: string }): Promise<User> {
    return request.post('/users', data)
  },

  // 更新用户
  updateUser(id: number, data: { username?: string; email?: string; role?: string }): Promise<User> {
    return request.put(`/users/${id}`, data)
  },

  // 删除用户
  deleteUser(id: number): Promise<void> {
    return request.delete(`/users/${id}`)
  },

  // 修改密码
  changePassword(data: ChangePasswordRequest): Promise<void> {
    return request.put('/me/password', data)
  },

  // 管理员重置用户密�?
  resetUserPassword(id: number, data: ResetPasswordRequest): Promise<void> {
    return request.put(`/users/${id}/password`, data)
  }
}


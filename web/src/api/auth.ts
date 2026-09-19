import request from './request'

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email?: string
  registration_code?: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  user: {
    id: string
    username: string
    email: string
    role: string
  }
}

export const authApi = {
  // 登录
  login(data: LoginRequest): Promise<LoginResponse> {
    return request.post('/auth/login', data)
  },

  // 注册
  register(data: RegisterRequest): Promise<LoginResponse> {
    return request.post('/auth/register', data)
  },

  // 刷新令牌
  refreshToken(): Promise<{ token: string }> {
    return request.post('/auth/refresh')
  }
}


import request from './request'

export interface RegistrationCode {
  id: string
  code: string
  status: string
  created_by: string
  created_at: string
  expires_at: string
  description?: string
  used_at?: string
  device_id?: string
}

export interface RegistrationCodeListParams {
  page?: number
  page_size?: number
  status?: string
}

export interface GenerateRegistrationCodeRequest {
  description?: string
  validity_days?: number
}

export const registrationCodeApi = {
  // 获取注册码列表
  getCodes(params?: RegistrationCodeListParams): Promise<{
    codes: RegistrationCode[]
    total: number
    page: number
    page_size: number
    total_pages: number
  }> {
    return request.get('/registration-codes', { params })
  },

  // 获取注册码详情
  getCode(id: string): Promise<RegistrationCode> {
    return request.get(`/registration-codes/${id}`)
  },

  // 生成注册码
  generateCode(data: GenerateRegistrationCodeRequest): Promise<RegistrationCode> {
    return request.post('/registration-codes', data)
  },

  // 删除注册码
  deleteCode(id: string): Promise<{ message: string }> {
    return request.delete(`/registration-codes/${id}`)
  },

  // 撤销注册码
  revokeCode(id: string): Promise<{ message: string }> {
    return request.post(`/registration-codes/${id}/revoke`)
  },

  // 重置注册码
  resetCode(id: string, data?: { extend_days?: number }): Promise<{ message: string }> {
    return request.post(`/registration-codes/${id}/reset`, data)
  },

  // 清理过期注册码
  cleanupExpiredCodes(): Promise<{ message: string; deleted_count: number }> {
    return request.post('/registration-codes/cleanup')
  }
}

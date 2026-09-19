import request from './request'

export interface Device {
  id: string
  name: string
  ip_address: string
  manufacturer: string
  android_version: string
  app_version: string
  battery_level?: number
  status: 'online' | 'offline' | 'error'
  last_seen: string
  created_at: string
}

export interface DeviceLog {
  id: string
  device_id: string
  device_name: string
  action: string
  details?: string
  remote_ip?: string
  created_at: string
}

export const deviceApi = {
  // 获取设备列表
  getDeviceList(params?: { page?: number; page_size?: number; status?: string }): Promise<{
    items: Device[]
    total: number
  }> {
    return request.get('/devices', { params })
  },

  // 获取设备详情
  getDevice(id: string): Promise<Device> {
    return request.get(`/devices/${id}`)
  },

  // 创建设备
  createDevice(data: { name: string; ip_address?: string; manufacturer?: string; android_version?: string; app_version?: string }): Promise<Device> {
    return request.post('/devices', data)
  },

  // 更新设备
  updateDevice(id: string, data: { name?: string }): Promise<Device> {
    return request.put(`/devices/${id}`, data)
  },

  // 删除设备
  deleteDevice(id: string): Promise<void> {
    return request.delete(`/devices/${id}`)
  },

  // 控制设备
  controlDevice(id: string, data: { command: string; parameters?: Record<string, any> }): Promise<void> {
    return request.post(`/devices/${id}/control`, data)
  },

  // 获取设备日志
  getDeviceLogs(id: string, params?: { page?: number; page_size?: number }): Promise<{
    items: DeviceLog[]
    total: number
  }> {
    return request.get(`/devices/${id}/logs`, { params })
  },

  // 发送设备心跳
  sendHeartbeat(deviceId: string, data: { status: 'online' | 'offline' | 'error'; timestamp?: string }): Promise<any> {
    return request.post(`/device/${deviceId}/heartbeat`, data)
  },

  // 获取LiveKit Token
  getLiveKitToken(id: string, role: 'viewer' | 'talkback' = 'viewer'): Promise<{
    token: string
    room_id: string
    server_url: string
  }> {
    return request.get(`/devices/${id}/livekit-token`, { params: { role } })
  },

  // 设置码率
  setBitrate(id: string, data: {
    target_kbps: number
    min_kbps?: number
    max_kbps?: number
    adaptive?: boolean
  }): Promise<void> {
    return request.post(`/devices/${id}/control`, {
      command: 'set_bitrate',
      parameters: data
    })
  },

  // 切换分辨率
  setResolution(id: string, data: {
    resolution: '480p' | '720p' | '1080p'
  }): Promise<void> {
    return request.post(`/devices/${id}/control`, {
      command: 'set_resolution',
      parameters: data
    })
  }
}


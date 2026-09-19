import request from './request'

export interface ActivityLog {
  id: string
  device_id: string
  device_name: string
  action: string
  details?: string
  user?: string
  created_at: string
}

export interface ActivityLogListResponse {
  items: ActivityLog[]
  total: number
}

export function getActivityLogs(params?: { 
  page?: number
  page_size?: number
  device_id?: string
  start_time?: string
  end_time?: string
}): Promise<ActivityLogListResponse> {
  return request.get('/logs', { params })
}

export function getDeviceLogs(deviceId: number, params?: { 
  page?: number
  page_size?: number
}): Promise<ActivityLogListResponse> {
  return request.get(`/devices/${deviceId}/logs`, { params })
}


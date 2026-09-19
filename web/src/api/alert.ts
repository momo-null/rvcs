import request from './request'

export interface Alert {
  id: string
  device_id: string
  rule_id?: string
  event_type: string
  level: 'info' | 'warning' | 'error' | 'critical'
  message?: string
  data?: Record<string, unknown>
  acknowledged: boolean
  acknowledged_by?: string
  acknowledged_at?: string
  created_at: string
}

export interface AlertStats {
  total: number
  unacknowledged: number
  acknowledged: number
  by_level: Record<string, number>
  by_device: Array<{
    device_id: string
    alert_count: number
  }>
}

export const alertApi = {
  getAlerts(params?: {
    page?: number
    page_size?: number
    acknowledged?: boolean
    level?: string
    device_id?: string
  }): Promise<{
    alerts: Alert[]
    total: number
  }> {
    return request.get('/alerts', { params })
  },

  getAlertStats(): Promise<AlertStats> {
    return request.get('/alerts/stats')
  },

  acknowledgeAlert(id: string): Promise<void> {
    return request.post(`/alerts/${id}/acknowledge`)
  },

  resolveAlert(id: string): Promise<void> {
    return request.post(`/alerts/${id}/resolve`)
  }
}

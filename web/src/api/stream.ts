import request from './request'

export interface Stream {
  stream_id: string
  device_id: number
  device_name?: string
  type: 'webrtc' | 'livekit'
  status: 'active' | 'inactive'
  url?: string
  created_at: string
}

export const streamApi = {
  // 获取所有流
  getAllStreams(): Promise<Stream[]> {
    return request.get('/streams')
  },

  // 获取流详�?
  getStream(streamId: string): Promise<Stream> {
    return request.get(`/streams/${streamId}`)
  },

  // 删除�?
  deleteStream(streamId: string): Promise<void> {
    return request.delete(`/streams/${streamId}`)
  }
}


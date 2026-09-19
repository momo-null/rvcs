import request from './request'

export interface LiveKitParticipantSummary {
  identity: string
  state: string
  tracks: number
}

export interface LiveKitRoomSummary {
  name: string
  created_at: number
  duration_sec: number
  participants: number
  publishers: number
  has_video: boolean
  has_talkback: boolean
  members: LiveKitParticipantSummary[]
}

export interface LiveKitOverview {
  service_status: 'up' | 'down'
  kpi: {
    active_rooms: number
    total_participants: number
    streaming_device_rooms: number
    talkback_sessions: number
  }
  rooms: LiveKitRoomSummary[]
}

export const monitorApi = {
  getLiveKitOverview(): Promise<LiveKitOverview> {
    return request.get('/monitor/livekit-overview')
  }
}



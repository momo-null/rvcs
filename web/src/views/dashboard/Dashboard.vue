<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #409eff">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalDevices }}</div>
              <div class="stat-label">Total Devices</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #67c23a">
              <el-icon><CircleCheck /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.onlineDevices }}</div>
              <div class="stat-label">Online Devices</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #e6a23c">
              <el-icon><Warning /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.pendingAlerts }}</div>
              <div class="stat-label">Pending Alerts</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #909399">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.totalUsers }}</div>
              <div class="stat-label">Total Users</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #5d7cff">
              <el-icon><Grid /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ livekit.kpi.active_rooms }}</div>
              <div class="stat-label">LiveKit Rooms</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #36b37e">
              <el-icon><UserFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ livekit.kpi.total_participants }}</div>
              <div class="stat-label">LiveKit Participants</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #00b8d9">
              <el-icon><VideoCamera /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ livekit.kpi.streaming_device_rooms }}</div>
              <div class="stat-label">Streaming Rooms</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background: #ff7849">
              <el-icon><Microphone /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ livekit.kpi.talkback_sessions }}</div>
              <div class="stat-label">Talkback Sessions</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>LiveKit Active Rooms</span>
              <el-tag :type="livekit.service_status === 'up' ? 'success' : 'danger'">
                {{ livekit.service_status === 'up' ? 'Service Up' : 'Service Down' }}
              </el-tag>
            </div>
          </template>

          <el-table :data="livekit.rooms" stripe>
            <el-table-column prop="name" label="Room" min-width="220" />
            <el-table-column prop="participants" label="Participants" width="100" />
            <el-table-column prop="publishers" label="Publishers" width="100" />
            <el-table-column prop="duration_sec" label="Duration" width="120">
              <template #default="{ row }">{{ formatDuration(row.duration_sec) }}</template>
            </el-table-column>
            <el-table-column label="Members" min-width="280">
              <template #default="{ row }">
                <el-space wrap>
                  <el-tag
                    v-for="member in row.members"
                    :key="`${row.name}-${member.identity}`"
                    size="small"
                    type="info"
                  >
                    {{ member.identity }}
                  </el-tag>
                </el-space>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>Recent Alerts</span>
              <el-button type="primary" link @click="$router.push('/alerts')">View All</el-button>
            </div>
          </template>
          <el-table :data="recentAlerts" stripe>
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="device_id" label="Device" />
            <el-table-column prop="event_type" label="Type" width="140" />
            <el-table-column prop="level" label="Level" width="100">
              <template #default="{ row }">
                <el-tag :type="getAlertLevelType(row.level)">{{ row.level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="Status" width="120">
              <template #default="{ row }">
                <el-tag :type="getAlertStatusType(row.acknowledged)">
                  {{ row.acknowledged ? 'acknowledged' : 'pending' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="Created At" width="180">
              <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import {
  CircleCheck,
  Grid,
  Microphone,
  Monitor,
  User,
  UserFilled,
  VideoCamera,
  Warning
} from '@element-plus/icons-vue'
import { deviceApi, type Device } from '@/api/device'
import { alertApi, type Alert } from '@/api/alert'
import { userApi } from '@/api/user'
import { monitorApi, type LiveKitOverview } from '@/api/monitor'
import dayjs from 'dayjs'

const REFRESH_INTERVAL_MS = 15000
let refreshTimer: number | undefined

const stats = ref({
  totalDevices: 0,
  onlineDevices: 0,
  pendingAlerts: 0,
  totalUsers: 0
})

const livekit = ref<LiveKitOverview>({
  service_status: 'down',
  kpi: {
    active_rooms: 0,
    total_participants: 0,
    streaming_device_rooms: 0,
    talkback_sessions: 0
  },
  rooms: []
})

const recentAlerts = ref<Alert[]>([])

const getAlertLevelType = (level: string) => {
  const types: Record<string, 'info' | 'warning' | 'danger'> = {
    info: 'info',
    warning: 'warning',
    error: 'danger',
    critical: 'danger'
  }
  return types[level] || 'info'
}

const getAlertStatusType = (acknowledged: boolean) => {
  return acknowledged ? 'info' : 'warning'
}

const formatTime = (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm:ss')

const formatDuration = (durationSec: number) => {
  const sec = Math.max(durationSec || 0, 0)
  const hours = Math.floor(sec / 3600)
  const minutes = Math.floor((sec % 3600) / 60)
  const seconds = sec % 60
  const hh = String(hours).padStart(2, '0')
  const mm = String(minutes).padStart(2, '0')
  const ss = String(seconds).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

const loadStats = async () => {
  try {
    const [devices, alertStats, users, alerts, lk] = await Promise.all([
      deviceApi.getDeviceList({ page_size: 1000 }),
      alertApi.getAlertStats(),
      userApi.getUserList({ page_size: 1000 }),
      alertApi.getAlerts({ page: 1, page_size: 10 }),
      monitorApi.getLiveKitOverview()
    ])

    stats.value.totalDevices = devices.total
    stats.value.onlineDevices = devices.items.filter((d: Device) => d.status === 'online').length
    stats.value.pendingAlerts = alertStats.unacknowledged
    stats.value.totalUsers = users.total
    recentAlerts.value = alerts.alerts
    livekit.value = lk
  } catch (error) {
    console.error('Failed to load dashboard stats:', error)
  }
}

onMounted(() => {
  loadStats()
  refreshTimer = window.setInterval(loadStats, REFRESH_INTERVAL_MS)
})

onUnmounted(() => {
  if (refreshTimer) {
    window.clearInterval(refreshTimer)
  }
})
</script>

<style scoped lang="scss">
.dashboard {
  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;
      gap: 20px;

      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;

        .el-icon {
          font-size: 30px;
        }
      }

      .stat-info {
        flex: 1;

        .stat-value {
          font-size: 28px;
          font-weight: bold;
          color: #303133;
          line-height: 1;
        }

        .stat-label {
          font-size: 14px;
          color: #909399;
          margin-top: 8px;
        }
      }
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>

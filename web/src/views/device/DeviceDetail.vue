<template>
  <div class="device-detail">
    <el-page-header @back="$router.back()">
      <template #content>
        <span>设备详情</span>
      </template>
    </el-page-header>

    <el-card style="margin-top: 20px" v-loading="loading">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>设备详情</span>
          <el-button v-if="isMobile" size="small" text @click="toggleDeviceDetails">
            {{ deviceDetailsExpanded ? '收起' : '展开' }}
          </el-button>
        </div>
      </template>
      <div v-show="!isMobile || deviceDetailsExpanded">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="设备ID">{{ device?.id }}</el-descriptions-item>
        <el-descriptions-item label="设备名称">{{ device?.name }}</el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ device?.ip_address || '-' }}</el-descriptions-item>
        <el-descriptions-item label="电池电量">{{ typeof device?.battery_level === 'number' ? `${device.battery_level}%` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="制造商">{{ device?.manufacturer || '-' }}</el-descriptions-item>
        <el-descriptions-item label="Android版本">{{ device?.android_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="应用版本">{{ device?.app_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="设备状态" :span="2">
          <el-tag :type="getStatusType(device?.status)">
            {{ getStatusText(device?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="最后在线时间" :span="2">
          {{ device?.last_seen ? formatTime(device.last_seen) : '从未在线' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ device?.created_at ? formatTime(device.created_at) : '' }}</el-descriptions-item>
      </el-descriptions>
      </div>
    </el-card>

    <!-- 摄像头控制 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <span>摄像头控制</span>
      </template>
      <div class="camera-control">
        <el-space>
          <el-select
            v-model="selectedResolution"
            placeholder="选择分辨率"
            style="width: 150px"
            :disabled="isCameraOn"
          >
            <el-option label="480p" value="480p" />
            <el-option label="720p" value="720p" />
            <el-option label="1080p" value="1080p" />
          </el-select>
          <el-button
            :type="isCameraOn ? 'danger' : 'primary'"
            :size="isMobile ? 'small' : 'default'"
            :loading="cameraControlling"
            @click="handleToggleCamera"
          >
            {{ isCameraOn ? '停止摄像头' : '启动摄像头' }}
          </el-button>
          <el-button
            :type="isTalkbackOn ? 'warning' : 'success'"
            :size="isMobile ? 'small' : 'default'"
            :loading="talkbackControlling"
            :disabled="!isTalkbackOn && (!isCameraOn || connectionStatus !== 'connected')"
            @click="toggleTalkback"
          >
            {{ isTalkbackOn ? '停止语音' : '开始语音' }}
          </el-button>
        </el-space>
        <el-space class="status-badge">
          <el-tag :type="isCameraOn ? 'success' : 'info'">
            {{ isCameraOn ? '已启动' : '已停止' }}
          </el-tag>
          <el-tag :type="isTalkbackOn ? 'success' : 'info'">
            {{ isTalkbackOn ? '语音中' : '未语音' }}
          </el-tag>
          <el-tag v-if="!isCameraOn" type="info">
            当前分辨率: {{ selectedResolution }}
          </el-tag>
          <el-tag v-if="cameraControlMessage" type="warning">
            {{ cameraControlMessage }}
          </el-tag>
        </el-space>
      </div>
    </el-card>

    <!-- 视频播放区域 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>实时视频</span>
        </div>
      </template>
      <el-space class="video-controls" wrap>
        <el-button size="small" text @click="toggleVideoRotation180">
          {{ videoRotationDeg === 180 ? '恢复方向' : '旋转180°' }}
        </el-button>
        <el-slider
          v-if="isMobile"
          v-model="videoHeight"
          :min="200"
          :max="800"
          :step="10"
          :marks="{ 200: '小', 400: '中', 600: '大', 800: '最大' }"
          style="width: 120px; margin-left: 10px;"
        />
      </el-space>
      <div class="video-container" ref="videoContainer" :style="{ height: `${videoHeight}px` }">
        <div v-if="!isCameraOn" class="video-placeholder">
          <el-icon size="48"><VideoCamera /></el-icon>
          <p>请先启动摄像头</p>
        </div>
        <div v-else-if="!streamReady" class="video-placeholder">
          <el-icon class="is-loading" size="48"><Loading /></el-icon>
          <p>正在连接摄像头...</p>
        </div>
        <div v-else-if="connectionStatus === 'connecting'" class="video-placeholder">
          <el-icon class="is-loading" size="48"><Loading /></el-icon>
          <p>正在建立视频连接...</p>
        </div>
        <video
          v-else-if="streamReady || connectionStatus === 'connected' || connectionStatus === 'idle'"
          ref="videoElement"
          autoplay
          playsinline
          muted
          class="video-element"
          :style="{ transform: `translate(-50%, -50%) rotate(${videoRotationDeg}deg)` }"
        />
        <div v-else class="video-placeholder">
          <el-icon size="48"><Warning /></el-icon>
          <p>连接失败</p>
          <el-button type="primary" @click="manualReconnectLiveKit">重新连接</el-button>
        </div>
      </div>
    </el-card>

    <!-- 流状态信息 -->
    <el-card style="margin-top: 20px" v-if="isCameraOn">
      <template #header>
        <span>流状态</span>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="流类型">LiveKit</el-descriptions-item>
        <el-descriptions-item label="分辨率">{{ selectedResolution }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStreamStatusType(streamStatus)">{{ getStreamStatusText(streamStatus) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="连接状态">
          <el-tag :type="getConnectionStatusType(connectionStatus)">
            {{ getConnectionStatusText(connectionStatus) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="网络质量">
          <el-tag :type="getNetworkQualityType(networkQuality)">
            {{ getNetworkQualityText(networkQuality) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="自适应码率">
          <el-tag :type="adaptiveBitrateEnabled ? 'success' : 'info'">
            {{ adaptiveBitrateEnabled ? '已启用' : '已禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">
          {{ streamStartTime ? formatTime(streamStartTime) : '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="码率">{{ currentBitrate }} kbps</el-descriptions-item>
      </el-descriptions>
    </el-card>

    <!-- 码率控制 -->
    <el-card style="margin-top: 20px" v-if="isCameraOn">
      <template #header>
        <span>码率控制</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="12">
          <div class="bitrate-control">
            <el-space>
              <span>手动设置码率:</span>
              <el-input-number
                v-model="targetBitrate"
                :min="100"
                :max="4000"
                :step="100"
                :disabled="adaptiveBitrateEnabled"
                style="width: 150px"
              />
              <span>kbps</span>
              <el-button
                type="primary"
                size="small"
                @click="applyBitrate"
                :disabled="adaptiveBitrateEnabled"
              >
                应用
              </el-button>
            </el-space>
          </div>
        </el-col>
        <el-col :span="12">
          <div class="bitrate-control">
            <el-space>
              <span>自适应码率:</span>
              <el-switch
                v-model="adaptiveBitrateEnabled"
                @change="toggleAdaptiveBitrate"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-space>
          </div>
        </el-col>
      </el-row>
      <el-divider />
      <div class="bitrate-info">
        <el-space>
          <el-tag type="info">当前码率: {{ currentBitrate }} kbps</el-tag>
          <el-tag type="info">最小码率: 300 kbps</el-tag>
          <el-tag type="info">最大码率: 4000 kbps</el-tag>
        </el-space>
      </div>
    </el-card>

    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span>设备日志</span>
          <el-button v-if="isMobile" size="small" text @click="toggleLogs">
            {{ logsExpanded ? '收起' : '展开' }}
          </el-button>
        </div>
      </template>
      <div v-show="!isMobile || logsExpanded">
      <el-table :data="logs" v-loading="loadingLogs" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="device_name" label="设备名称" width="150" />
        <el-table-column prop="action" label="操作" width="120">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)">{{ row.action }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="details" label="详情" show-overflow-tooltip />
        <el-table-column prop="remote_ip" label="IP地址" width="140" />
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoCamera, Loading, Warning } from '@element-plus/icons-vue'
import { deviceApi } from '@/api/device'
import type { Device, DeviceLog } from '@/api/device'
import dayjs from 'dayjs'
import { Room, RoomEvent, Track } from 'livekit-client'
import { getDeviceAutoRefreshIntervalMs } from '@/utils/settings'

// 窗口大小变化监听
const handleResize = () => {
  checkMobile()
}

const route = useRoute()
const loading = ref(false)
const loadingLogs = ref(false)
const device = ref<Device | null>(null)
const logs = ref<DeviceLog[]>([])
let deviceRefreshTimer: number | null = null
let logsRefreshTimer: number | null = null
const cameraControlling = ref(false)
const isCameraOn = ref(false)
const cameraControlMessage = ref('')
const selectedResolution = ref('720p')
const CAMERA_ON_ACK_TIMEOUT_MS = 12000
const HEARTBEAT_STALE_SECONDS = 120
let cameraOnAckTimer: number | null = null
const pendingCameraOnCommand = ref(false)

// 视频播放器引用
const videoElement = ref<HTMLVideoElement>()
const videoContainer = ref<HTMLDivElement>()

// 移动端检测和视频高度控制
const isMobile = ref(false)
const deviceDetailsExpanded = ref(false)
const logsExpanded = ref(false)
const videoHeight = ref(480) // 默认视频高度
const videoRotationDeg = ref(0)
const VIDEO_PREFS_KEY_PREFIX = 'device_video_view_prefs'
const videoPrefsLoaded = ref(false)
const hasVideoHeightPreference = ref(false)

// 流状态
const streamStatus = ref<'inactive' | 'starting' | 'active' | 'stopping'>('inactive')
const connectionStatus = ref<'idle' | 'connecting' | 'connected' | 'failed'>('idle')
const streamStartTime = ref<string | null>(null)
const streamReady = ref(false)
const isTalkbackOn = ref(false)
const talkbackControlling = ref(false)
const isFlashlightOn = ref(false)
const flashlightControlling = ref(false)

// 网络质量和码率控制
const networkQuality = ref<'poor' | 'good' | 'excellent'>('good')
const adaptiveBitrateEnabled = ref(true)
const targetBitrate = ref(1200)  // 目标码率 (kbps)
const currentBitrate = ref(1200)  // 当前码率 (kbps)

// WebSocket
let ws: WebSocket | null = null
const wsReconnectAttempts = ref(0)
const MAX_RECONNECT_ATTEMPTS = 5

const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

const toggleDeviceDetails = () => {
  deviceDetailsExpanded.value = !deviceDetailsExpanded.value
}

const toggleLogs = () => {
  logsExpanded.value = !logsExpanded.value
}

const toggleVideoRotation180 = () => {
  videoRotationDeg.value = videoRotationDeg.value === 180 ? 0 : 180
}

const getVideoPrefsKey = () => {
  const deviceId = String(route.params.id || 'unknown')
  return `${VIDEO_PREFS_KEY_PREFIX}:${deviceId}`
}

const loadVideoViewPreferences = () => {
  try {
    const raw = localStorage.getItem(getVideoPrefsKey())
    if (!raw) {
      videoPrefsLoaded.value = true
      return
    }

    const prefs = JSON.parse(raw) as { height?: number; rotation?: number }

    if (typeof prefs.height === 'number' && prefs.height >= 200 && prefs.height <= 800) {
      videoHeight.value = prefs.height
      hasVideoHeightPreference.value = true
    }

    if (prefs.rotation === 0 || prefs.rotation === 180) {
      videoRotationDeg.value = prefs.rotation
    }
  } catch (error) {
    console.warn('Failed to load video view preferences', error)
  } finally {
    videoPrefsLoaded.value = true
  }
}

const saveVideoViewPreferences = () => {
  if (!videoPrefsLoaded.value) return

  const prefs = {
    height: videoHeight.value,
    rotation: videoRotationDeg.value
  }

  localStorage.setItem(getVideoPrefsKey(), JSON.stringify(prefs))
}

const clearCameraOnAckTimer = () => {
  if (cameraOnAckTimer !== null) {
    window.clearTimeout(cameraOnAckTimer)
    cameraOnAckTimer = null
  }
}

const isHeartbeatLikelyStale = () => {
  if (!device.value?.last_seen) return true
  const seconds = dayjs().diff(dayjs(device.value.last_seen), 'second')
  return seconds > HEARTBEAT_STALE_SECONDS
}

const getStatusType = (status?: string) => {
  const types: Record<string, any> = {
    online: 'success',
    offline: 'info',
    error: 'danger'
  }
  return types[status || ''] || 'info'
}

const getStatusText = (status?: string) => {
  const texts: Record<string, string> = {
    online: '在线',
    offline: '离线',
    error: '错误'
  }
  return texts[status || ''] || '未知'
}

const getActionType = (action: string) => {
  const types: Record<string, any> = {
    'register': 'success',
    'heartbeat': 'info',
    'disconnect': 'warning',
    'error': 'danger'
  }
  return types[action] || 'info'
}

const getStreamStatusType = (status: string) => {
  const types: Record<string, any> = {
    inactive: 'info',
    starting: 'warning',
    active: 'success',
    stopping: 'warning'
  }
  return types[status] || 'info'
}

const getStreamStatusText = (status: string) => {
  const texts: Record<string, string> = {
    inactive: '未启动',
    starting: '启动中',
    active: '活跃',
    stopping: '停止中'
  }
  return texts[status] || '未知'
}

const getConnectionStatusType = (status: string) => {
  const types: Record<string, any> = {
    idle: 'info',
    connecting: 'warning',
    connected: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const getConnectionStatusText = (status: string) => {
  const texts: Record<string, string> = {
    idle: '未连接',
    connecting: '连接中',
    connected: '已连接',
    failed: '连接失败'
  }
  return texts[status] || '未知'
}

const getNetworkQualityType = (quality: string) => {
  const types: Record<string, any> = {
    poor: 'danger',
    good: 'success',
    excellent: 'success'
  }
  return types[quality] || 'info'
}

const getNetworkQualityText = (quality: string) => {
  const texts: Record<string, string> = {
    poor: '差',
    good: '良好',
    excellent: '优秀'
  }
  return texts[quality] || '未知'
}

// 应用手动码率
const applyBitrate = async () => {
  if (!device.value || adaptiveBitrateEnabled.value) return

  try {
    await deviceApi.setBitrate(device.value.id, {
      target_kbps: targetBitrate.value,
      adaptive: false
    })
    ElMessage.success(`码率已设置为 ${targetBitrate.value} kbps`)
    currentBitrate.value = targetBitrate.value
  } catch (error: any) {
    console.error('Failed to set bitrate:', error)
    ElMessage.error(error?.message || '设置码率失败')
  }
}

// 切换自适应码率
const toggleAdaptiveBitrate = async (val: string | number | boolean) => {
  if (!device.value) return

  // 转换为布尔值
  const enabled = val === true || val === 'true'

  try {
    await deviceApi.setBitrate(device.value.id, {
      target_kbps: 1200,  // 默认目标码率
      adaptive: enabled
    })
    ElMessage.success(enabled ? '自适应码率已启用' : '自适应码率已禁用')
  } catch (error: any) {
    console.error('Failed to toggle adaptive bitrate:', error)
    ElMessage.error(error?.message || '切换自适应码率失败')
  }
}

// 初始化WebSocket连接
const initWebSocket = () => {
  console.log('========== 初始化 WebSocket 连接 ==========')

  const token = localStorage.getItem('token')
  if (!token) {
    console.error('❌ 未登录，无法连接 WebSocket')
    ElMessage.error('未登录')
    return
  }

  console.log('Token 存在，长度:', token.length)

  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/api/v1/web/ws?token=${token}`

  console.log('WebSocket URL:', wsUrl)

  try {
    console.log('创建 WebSocket 连接...')
    ws = new WebSocket(wsUrl)
    wsReconnectAttempts.value = 0

    ws.onopen = () => {
      console.log('✅ WebSocket 连接成功')
      console.log('连接 URL:', wsUrl)
      wsReconnectAttempts.value = 0
    }

    ws.onmessage = (event) => {
      console.log('========== 收到 WebSocket 消息 ==========')
      console.log('Raw message:', event.data)
      try {
        const message = JSON.parse(event.data)
        handleWebSocketMessage(message)
      } catch (error) {
        console.error('❌ 解析 WebSocket 消息失败:', error)
        console.error('Raw data:', event.data)
      }
    }

    ws.onclose = () => {
      console.log('⚠️ WebSocket 连接已断开')
      scheduleReconnect()
    }

    ws.onerror = (error) => {
      console.error('❌ WebSocket 错误:', error)
    }
  } catch (error) {
    console.error('❌ 创建 WebSocket 失败:', error)
    scheduleReconnect()
  }
}

// 处理WebSocket消息
const handleWebSocketMessage = (message: any) => {
  console.log('========== 收到 WebSocket 消息 ==========')
  console.log('Message:', message)

  const type = message.type
  const data = message.data ?? message

  console.log('Message type:', type)
  console.log('Message data:', data)

  switch (type) {
    case 'camera_state_changed':
      console.log('处理摄像头状态变更')
      handleCameraStateChanged(data)
      break
    case 'command_response':
      console.log('处理命令响应')
      handleCommandResponse(data)
      break
    case 'device_status_changed':
      console.log('处理设备状态变更')
      handleDeviceStatusChanged(data)
      break
    default:
      console.log('❌ 未知消息类型:', type)
  }
}

// 处理摄像头状态变更
const handleCameraStateChanged = (data: any) => {
  console.log('========== 摄像头状态变更 ==========')
  console.log('Data:', data)
  console.log('State:', data.state)

  // 只处理摄像头开启状态
  if (data.state === 'on') {
    pendingCameraOnCommand.value = false
    clearCameraOnAckTimer()
    console.log('✅ 摄像头已开启')
    isCameraOn.value = true
    streamStatus.value = 'active'
    streamStartTime.value = new Date().toISOString()

    // 延迟连接 LiveKit，给 Android 端足够时间发布轨道
    console.log('等待 3 秒后连接 LiveKit...')
    setTimeout(() => {
      if (streamStatus.value === 'active' && shouldAutoReconnect) {
        console.log('准备连接 LiveKit...')
        const deviceId = route.params.id as string
        connectLiveKit(deviceId)
      }
    }, 3000)
  } else if (data.state === 'off') {
    pendingCameraOnCommand.value = false
    clearCameraOnAckTimer()
    console.log('📷 摄像头已关闭')
    isCameraOn.value = false
    streamStatus.value = 'inactive'
    // 不需要断开 LiveKit，因为 handleStopCamera 已经处理了
  } else if (data.state === 'error') {
    pendingCameraOnCommand.value = false
    clearCameraOnAckTimer()
    console.log('❌ 摄像头错误:', data.message)
    ElMessage.error(data.message || '摄像头发生错误')
  }
}

// 处理命令响应
const handleCommandResponse = (data: any) => {
  console.log('========== 命令响应 ==========')
  console.log('Data:', data)
  const command = data?.command
  const successValue = data?.success

  if (command === 'camera_on') {
    if (successValue === true) {
      pendingCameraOnCommand.value = false
      clearCameraOnAckTimer()
      if (!isCameraOn.value) {
        cameraControlMessage.value = '命令已确认，等待设备推流...'
      }
    } else if (successValue === false) {
      pendingCameraOnCommand.value = false
      clearCameraOnAckTimer()
      streamStatus.value = 'inactive'
      cameraControlMessage.value = '启动失败'
      ElMessage.error(data?.message || '设备拒绝启动摄像头')
    } else {
      // Unknown command response shape, keep waiting for camera_state_changed/timeout.
      console.log('camera_on response without explicit success field, keep waiting')
    }
    return
  }

  if (command === 'flashlight_on') {
    if (successValue === true) {
      isFlashlightOn.value = true
    } else if (successValue === false) {
      isFlashlightOn.value = false
      ElMessage.error(data?.message || '打开手电筒失败')
    }
    return
  }

  if (command === 'flashlight_off') {
    if (successValue === true) {
      isFlashlightOn.value = false
    } else if (successValue === false) {
      ElMessage.error(data?.message || '关闭手电筒失败')
    }
  }
}

// 处理设备状态变更
const handleDeviceStatusChanged = (data: any) => {
  const status = data.status
  console.log('Device status changed:', status)

  if (device.value) {
    device.value.status = status
  }
  loadDevice(true)
  loadLogs(true)
}

// 处理流启动
const handleStreamStarted = async (data: any) => {
  console.log('Stream started:', data)
  streamStatus.value = 'active'
  streamStartTime.value = new Date().toISOString()
  ElMessage.success('视频流已启动，正在建立LiveKit连接...')

  // 连接LiveKit
  const deviceId = route.params.id as string
  await connectLiveKit(deviceId)
}

// WebSocket重连
const scheduleReconnect = () => {
  if (wsReconnectAttempts.value >= MAX_RECONNECT_ATTEMPTS) {
    console.log('Max reconnect attempts reached')
    return
  }

  wsReconnectAttempts.value++
  const delay = Math.min(1000 * Math.pow(2, wsReconnectAttempts.value), 30000)

  console.log(`Reconnecting in ${delay}ms...`)
  setTimeout(() => {
    if (!ws || ws.readyState === WebSocket.CLOSED) {
      initWebSocket()
    }
  }, delay)
}

// 关闭WebSocket
const closeWebSocket = () => {
  if (ws) {
    ws.close()
    ws = null
  }
}

// LiveKit Room 实例
let liveKitRoom: Room | null = null

// LiveKit 重连控制
const liveKitReconnectAttempts = ref(0)
const MAX_LIVEKIT_RECONNECT_ATTEMPTS = 5
let liveKitReconnectTimer: number | null = null
let shouldAutoReconnect = true  // 是否允许自动重连（手动停止时为 false）
let isLiveKitConnecting = false
let liveKitSessionId = 0

const clearLiveKitReconnectTimer = () => {
  if (liveKitReconnectTimer) {
    clearTimeout(liveKitReconnectTimer)
    liveKitReconnectTimer = null
  }
}

// 连接LiveKit
const connectLiveKit = async (deviceId: string) => {
  if (isLiveKitConnecting) {
    console.log('Skip connectLiveKit: connection attempt already running')
    return
  }
  isLiveKitConnecting = true
  const sessionId = ++liveKitSessionId

  console.log('========== 开始连接 LiveKit ==========')
  console.log('Device ID:', deviceId)
  console.log('Video element exists:', videoElement.value !== null)

  try {
    connectionStatus.value = 'connecting'
    clearLiveKitReconnectTimer()

    if (liveKitRoom) {
      try {
        liveKitRoom.disconnect()
      } catch (error) {
        console.warn('Disconnect old room before reconnect failed:', error)
      } finally {
        liveKitRoom = null
      }
    }

    console.log('正在获取 LiveKit Token...')

    // 获取LiveKit Token
    const tokenData = await deviceApi.getLiveKitToken(deviceId, 'talkback')
    const { token, server_url, room_id } = tokenData

    console.log('✅ LiveKit Token 获取成功')
    console.log('  - Server URL:', server_url)
    console.log('  - Room ID:', room_id)
    console.log('  - Token length:', token?.length)

    // 创建LiveKit Room
    console.log('创建 LiveKit Room 实例...')
    liveKitRoom = new Room()
    console.log('✅ Room 实例创建成功')

    // 监听参与者连接
    liveKitRoom.on(RoomEvent.ParticipantConnected, (participant) => {
      console.log('========== 参与者已连接 ==========')
      console.log('Participant identity:', participant.identity)
      console.log('Participant name:', participant.name)
      console.log('Participant track publications:', participant.trackPublications.size)

      // 订阅参与者的所有轨道
      participant.trackPublications.forEach((publication) => {
        console.log('  - Track:', publication.kind, publication.trackSid, 'Subscribed:', publication.isSubscribed)
        if (!publication.isSubscribed) {
          console.log('    订阅轨道:', publication.trackSid)
          publication.setSubscribed(true)
        }
      })
    })

    liveKitRoom.on(RoomEvent.TrackSubscribed, (track, publication, participant) => {
      console.log('========== 轨道已订阅 ==========')
      console.log('Track kind:', track.kind)
      console.log('Track SID:', publication.trackSid)
      console.log('Participant:', participant.identity)
      console.log('Video element exists:', videoElement.value !== null)

      if (track.kind === 'video') {
        if (videoElement.value) {
          console.log('✅ 附加视频元素到 DOM')
          // 附加视频元素
          track.attach(videoElement.value)
          streamReady.value = true
          connectionStatus.value = 'connected'
          ElMessage.success('LiveKit视频连接成功')
        } else {
          console.log('⚠️ 视频元素不存在，等待 DOM 更新')
          // 延迟附加，等待 Vue 更新 DOM
          setTimeout(() => {
            if (videoElement.value) {
              console.log('延迟附加视频元素')
              track.attach(videoElement.value)
              streamReady.value = true
              connectionStatus.value = 'connected'
              ElMessage.success('LiveKit视频连接成功')
            }
          }, 100)
        }
      } else if (track.kind === 'audio') {
        console.log('✅ 播放音频轨道')
        // 播放音频
        track.attach()
      }
    })

    liveKitRoom.on(RoomEvent.TrackUnsubscribed, (track) => {
      console.log('========== 轨道已取消订阅 ==========')
      console.log('Track kind:', track.kind)
      if (videoElement.value && track.kind === 'video') {
        track.detach(videoElement.value)
        videoElement.value.srcObject = null
        streamReady.value = false
      }
    })

    liveKitRoom.on(RoomEvent.Disconnected, () => {
      if (sessionId !== liveKitSessionId) return
      console.log('========== LiveKit 房间已断开 ==========')
      connectionStatus.value = 'idle'
      isTalkbackOn.value = false
      // 触发自动重连
      if (shouldAutoReconnect) {
        scheduleLiveKitReconnect(deviceId)
      }
    })

    liveKitRoom.on(RoomEvent.ConnectionStateChanged, (state) => {
      if (sessionId !== liveKitSessionId) return
      console.log('========== 连接状态变更 ==========')
      console.log('New state:', state)
      console.log('shouldAutoReconnect:', shouldAutoReconnect)

      if (state === 'connected') {
        console.log('✅ LiveKit 连接成功')
        connectionStatus.value = 'connected'
        streamReady.value = true
        // 连接成功，重置重连计数
        liveKitReconnectAttempts.value = 0
        clearLiveKitReconnectTimer()

        // 检查是否有远程参与者
        if (liveKitRoom) {
          const remoteParticipants = Array.from(liveKitRoom.remoteParticipants.values())
          console.log('远程参与者数量:', remoteParticipants.length)

          if (remoteParticipants.length === 0) {
            console.log('⚠️ 没有远程参与者，等待 Android 端加入...')
            return
          }

          // 订阅所有参与者的轨道
          remoteParticipants.forEach((participant) => {
            console.log('  - Participant:', participant.identity, 'Track publications:', participant.trackPublications.size)
            if (participant.trackPublications.size === 0) {
              console.log('  ⚠️ 参与者还没有发布轨道，等待...')
              return
            }
            participant.trackPublications.forEach((publication) => {
              console.log('    - Track:', publication.kind, publication.trackSid, 'Subscribed:', publication.isSubscribed)
              if (!publication.isSubscribed) {
                console.log('      订阅轨道:', publication.trackSid)
                publication.setSubscribed(true)
              }
            })
          })
        }
      } else if (state === 'disconnected' || state === 'reconnecting') {
        console.log('❌ 连接断开，状态:', state)
        connectionStatus.value = 'failed'
        streamReady.value = false
        isTalkbackOn.value = false
        // 触发自动重连
        if (shouldAutoReconnect) {
          console.log('触发自动重连...')
          scheduleLiveKitReconnect(deviceId)
        } else {
          console.log('不触发自动重连（shouldAutoReconnect = false）')
        }
      }
    })

    console.log('========== 连接到 LiveKit 房间 ==========')
    console.log('  - Server:', server_url)
    console.log('  - Room:', room_id)
    console.log('  - AutoSubscribe: true')

    // 连接到LiveKit房间，指定房间名称
    await liveKitRoom.connect(server_url, token, {
      autoSubscribe: true,
    })

    console.log('✅ LiveKit 房间连接成功!')
    console.log('  - Room name:', room_id)

    // 检查并订阅现有参与者的轨道
    const remoteParticipants = Array.from(liveKitRoom.remoteParticipants.values())
    console.log('========== 连接后检查参与者 ==========')
    console.log('远程参与者数量:', remoteParticipants.length)

    remoteParticipants.forEach((participant) => {
      console.log('  - Participant:', participant.identity)
      console.log('    Name:', participant.name)
      console.log('    Track publications count:', participant.trackPublications.size)
      participant.trackPublications.forEach((publication) => {
        console.log('      - Track:', publication.kind, publication.trackSid, 'Subscribed:', publication.isSubscribed)
        if (!publication.isSubscribed) {
          console.log('        订阅轨道:', publication.trackSid)
          publication.setSubscribed(true)
        }
      })
    })

    console.log('========== LiveKit 连接完成 ==========')
  } catch (error: any) {
    console.error('========== LiveKit 连接失败 ==========')
    console.error('Error:', error)
    console.error('Error message:', error?.message)
    console.error('Error stack:', error?.stack)
    connectionStatus.value = 'failed'
    ElMessage.error(error?.message || 'LiveKit连接失败')

    // 连接失败，尝试重连
    if (shouldAutoReconnect) {
      scheduleLiveKitReconnect(deviceId)
    }
  } finally {
    if (sessionId === liveKitSessionId) {
      isLiveKitConnecting = false
    }
  }
}

// LiveKit 重连调度（指数退避算法）
const scheduleLiveKitReconnect = (deviceId: string) => {
  if (connectionStatus.value === 'connected') {
    return
  }

  if (liveKitReconnectTimer) {
    return
  }

  // 取消已存在的重连定时器
  clearLiveKitReconnectTimer()

  // 检查重连次数限制
  if (liveKitReconnectAttempts.value >= MAX_LIVEKIT_RECONNECT_ATTEMPTS) {
    console.log('Max LiveKit reconnect attempts reached')
    ElMessage.error('LiveKit重连失败，请手动点击重新连接')
    connectionStatus.value = 'failed'
    return
  }

  // 增加重连计数
  liveKitReconnectAttempts.value++

  // 指数退避计算：1s, 2s, 4s, 8s, 16s, 最大30s
  const delay = Math.min(1000 * Math.pow(2, liveKitReconnectAttempts.value), 30000)

  console.log(`LiveKit reconnecting in ${delay}ms (attempt ${liveKitReconnectAttempts.value}/${MAX_LIVEKIT_RECONNECT_ATTEMPTS})`)

  // 更新UI显示重连状态
  connectionStatus.value = 'connecting'
  ElMessage.warning(`LiveKit连接断开，${delay / 1000}秒后自动重连...`)

  // 设置重连定时器
  liveKitReconnectTimer = window.setTimeout(async () => {
    liveKitReconnectTimer = null
    if (shouldAutoReconnect) {
      console.log('Attempting LiveKit reconnection...')
      await connectLiveKit(deviceId)
    }
  }, delay)
}

// 手动重新连接LiveKit
const manualReconnectLiveKit = async () => {
  if (!device.value) return

  try {
    connectionStatus.value = 'connecting'

    // 启用自动重连并重置计数
    shouldAutoReconnect = true
    liveKitReconnectAttempts.value = 0

    // 取消现有的重连定时器
    clearLiveKitReconnectTimer()

    // 连接LiveKit
    await connectLiveKit(device.value.id)
  } catch (error: any) {
    console.error('Manual reconnect failed:', error)
    ElMessage.error('手动重连失败')
  }
}

const startTalkback = async () => {
  if (!liveKitRoom || connectionStatus.value !== 'connected' || isTalkbackOn.value) return

  try {
    talkbackControlling.value = true
    await liveKitRoom.localParticipant.setMicrophoneEnabled(true)
    isTalkbackOn.value = true
    ElMessage.success('语音已开启')
  } catch (error: any) {
    console.error('Failed to start talkback:', error)
    ElMessage.error(error?.message || '开启语音失败')
  } finally {
    talkbackControlling.value = false
  }
}

const stopTalkback = async () => {
  if (!liveKitRoom || !isTalkbackOn.value) {
    isTalkbackOn.value = false
    return
  }

  try {
    talkbackControlling.value = true
    await liveKitRoom.localParticipant.setMicrophoneEnabled(false)
  } catch (error: any) {
    console.warn('Failed to stop talkback cleanly:', error)
  } finally {
    isTalkbackOn.value = false
    talkbackControlling.value = false
  }
}

const toggleTalkback = async () => {
  if (isTalkbackOn.value) {
    await stopTalkback()
    return
  }
  await startTalkback()
}

const toggleFlashlight = async () => {
  if (!device.value) return
  try {
    flashlightControlling.value = true
    const command = isFlashlightOn.value ? 'flashlight_off' : 'flashlight_on'
    await deviceApi.controlDevice(device.value.id, { command })
  } catch (error: any) {
    console.error('Failed to toggle flashlight:', error)
    ElMessage.error(error?.message || '手电筒操作失败')
  } finally {
    flashlightControlling.value = false
  }
}

// 断开LiveKit连接
const disconnectLiveKit = () => {
  console.log('========== 断开 LiveKit 连接 ==========')

  // 标记为手动断开，不触发自动重连
  shouldAutoReconnect = false

  // 取消重连定时器
  clearLiveKitReconnectTimer()

  // 重置重连计数
  liveKitReconnectAttempts.value = 0

  // 清理视频元素
  if (videoElement.value) {
    videoElement.value.srcObject = null
  }

  // 断开房间连接
  if (liveKitRoom) {
    // Best effort: close local mic publication before disconnect.
    liveKitRoom.localParticipant.setMicrophoneEnabled(false).catch((error: any) => {
      console.warn('Failed to disable mic before disconnect:', error)
    })
    isTalkbackOn.value = false
    console.log('断开 LiveKit 房间...')
    liveKitRoom.disconnect()
    liveKitRoom = null
  }

  liveKitSessionId++
  isLiveKitConnecting = false

  // 重置状态
  connectionStatus.value = 'idle'
  streamReady.value = false

  console.log('✅ LiveKit 连接已断开')
}

// 视频播放器事件处理
const handleVideoConnected = () => {
  console.log('Video player connected')
  connectionStatus.value = 'connected'
}

const handleVideoDisconnected = () => {
  console.log('Video player disconnected')
  connectionStatus.value = 'idle'
}

const handleVideoError = (error: string) => {
  console.error('Video player error:', error)
  connectionStatus.value = 'failed'
  ElMessage.error(error)
}

const handleStartCamera = async () => {
  console.log('========== 启动摄像头 ==========')
  console.log('Device ID:', device.value?.id)
  console.log('Device Name:', device.value?.name)
  console.log('Selected Resolution:', selectedResolution.value)

  if (!device.value) {
    console.error('❌ 设备不存在')
    return
  }

  try {
    cameraControlling.value = true
    cameraControlMessage.value = '正在启动...'
    streamStatus.value = 'starting'
    pendingCameraOnCommand.value = true
    clearCameraOnAckTimer()

    if (device.value.status !== 'online' || isHeartbeatLikelyStale()) {
      ElMessage.warning('设备心跳可能异常，已继续尝试下发开摄像头命令')
    }

    console.log('正在发送 camera_on 命令到设备...')

    // 启用自动重连
    shouldAutoReconnect = true
    liveKitReconnectAttempts.value = 0

    await deviceApi.controlDevice(device.value.id, {
      command: 'camera_on',
      parameters: {
        resolution: selectedResolution.value
      }
    })

    console.log('✅ camera_on 命令发送成功，等待设备响应...')
    cameraControlMessage.value = '命令已发送'
    ElMessage.success('摄像头启动命令已发送，请等待设备响应')

    cameraOnAckTimer = window.setTimeout(() => {
      if (!pendingCameraOnCommand.value || isCameraOn.value) return
      pendingCameraOnCommand.value = false
      streamStatus.value = 'inactive'
      cameraControlMessage.value = '启动超时'
      ElMessage.error('设备启动超时，请检查设备连接后重试')
    }, CAMERA_ON_ACK_TIMEOUT_MS)

    setTimeout(() => {
      cameraControlMessage.value = ''
    }, 3000)
  } catch (error: any) {
    console.error('❌ 摄像头启动失败')
    console.error('Error:', error)
    console.error('Error message:', error?.message)
    pendingCameraOnCommand.value = false
    clearCameraOnAckTimer()
    cameraControlMessage.value = '启动失败'
    streamStatus.value = 'inactive'
    ElMessage.error(error?.message || '摄像头启动失败')
    isCameraOn.value = false
    setTimeout(() => {
      cameraControlMessage.value = ''
    }, 3000)
  } finally {
    cameraControlling.value = false
  }
}

const handleStopCamera = async () => {
  if (!device.value) return

  try {
    cameraControlling.value = true
    pendingCameraOnCommand.value = false
    clearCameraOnAckTimer()
    cameraControlMessage.value = '正在停止...'
    streamStatus.value = 'stopping'
    await stopTalkback()

    // 先发送命令给 Android 端停止摄像头
    await deviceApi.controlDevice(device.value.id, { command: 'camera_off' })

    // 等待 4 秒让 Android 端停止摄像头、unpublish 轨道并断开 LiveKit 连接
    // 总等待时间 = 2秒禁用 + 2秒断开 = 4秒，再加上 buffer 时间
    await new Promise(resolve => setTimeout(resolve, 4000))

    // 断开前端的 LiveKit 连接
    disconnectLiveKit()

    // 重置状态
    isCameraOn.value = false
    connectionStatus.value = 'idle'
    streamReady.value = false
    cameraControlMessage.value = '停止成功'
    ElMessage.success('摄像头停止成功')

    setTimeout(() => {
      cameraControlMessage.value = ''
    }, 3000)
  } catch (error: any) {
    cameraControlMessage.value = '停止失败'
    streamStatus.value = 'active'
    ElMessage.error(error?.message || '摄像头停止失败')
    isCameraOn.value = true
    setTimeout(() => {
      cameraControlMessage.value = ''
    }, 3000)
  } finally {
    cameraControlling.value = false
  }
}

const handleToggleCamera = async () => {
  if (cameraControlling.value) return
  if (isCameraOn.value) {
    await handleStopCamera()
  } else {
    await handleStartCamera()
  }
}

const loadDevice = async (silent = false) => {
  console.log('========== 加载设备详情 ==========')
  const id = route.params.id as string
  console.log('Device ID from route:', id)

  try {
    if (!silent) loading.value = true
    console.log('正在调用 API 获取设备信息...')
    device.value = await deviceApi.getDevice(id)
    console.log('✅ 设备信息加载成功')
    console.log('Device:', device.value)
  } catch (error) {
    console.error('❌ 加载设备详情失败:', error)
    if (!silent) ElMessage.error('加载设备详情失败')
  } finally {
    if (!silent) loading.value = false
  }
}

const loadLogs = async (silent = false) => {
  try {
    if (!silent) loadingLogs.value = true
    const id = route.params.id as string
    const result = await deviceApi.getDeviceLogs(id, { page: 1, page_size: 50 })
    logs.value = result.items || []
  } catch (error) {
    if (!silent) ElMessage.error('加载设备日志失败')
  } finally {
    if (!silent) loadingLogs.value = false
  }
}

const startAutoRefresh = () => {
  stopAutoRefresh()

  deviceRefreshTimer = window.setInterval(() => {
    if (document.visibilityState !== 'visible') return
    loadDevice(true)
  }, getDeviceAutoRefreshIntervalMs())

  logsRefreshTimer = window.setInterval(() => {
    if (document.visibilityState !== 'visible') return
    loadLogs(true)
  }, getDeviceAutoRefreshIntervalMs())
}

const stopAutoRefresh = () => {
  if (deviceRefreshTimer !== null) {
    window.clearInterval(deviceRefreshTimer)
    deviceRefreshTimer = null
  }
  if (logsRefreshTimer !== null) {
    window.clearInterval(logsRefreshTimer)
    logsRefreshTimer = null
  }
}

const handleVisibilityChange = () => {
  if (document.visibilityState === 'visible') {
    loadDevice(true)
    loadLogs(true)
  }
}

onMounted(() => {
  console.log('========== DeviceDetail 组件已挂载 ==========')
  console.log('Route params:', route.params)

  // 检测移动端
  checkMobile()
  loadVideoViewPreferences()

  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)

  loadDevice()
  loadLogs()
  initWebSocket()
  startAutoRefresh()
  document.addEventListener('visibilitychange', handleVisibilityChange)

  console.log('组件初始化完成')
})

// 检测是否为移动端设备
const checkMobile = () => {
  const width = window.innerWidth
  const isTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0
  const userAgent = navigator.userAgent
  const isMobileUA = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(userAgent)

  isMobile.value = isTouch && (isMobileUA || width < 1024)

  // 根据设备类型设置默认视频高度
  // 如果用户已有偏好则不覆盖
  if (!hasVideoHeightPreference.value) {
    if (isMobile.value) {
      videoHeight.value = width < 480 ? 280 : 320
    } else {
      videoHeight.value = 480
    }
  }

  console.log('Device detection:', { isMobile: isMobile.value, width })
}

onBeforeUnmount(async () => {
  stopAutoRefresh()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  // 移除窗口大小监听
  window.removeEventListener('resize', handleResize)
  await stopTalkback()

  // 如果摄像头正在运行，先停止
  if (isCameraOn.value && device.value) {
    try {
      console.log('========== 离开页面，停止摄像头 ==========')
      cameraControlling.value = true

      // 1. 先发送停止命令给 Android
      await deviceApi.controlDevice(device.value.id, { command: 'camera_off' })
      console.log('✅ 停止命令已发送')

      // 2. 等待 Android 端停止推流（给足够时间 unpublish 轨道、停止硬件、断开连接）
      // 总等待时间 = 2秒禁用 + 2秒断开 = 4秒，再加上 buffer 时间
      await new Promise(resolve => setTimeout(resolve, 4000))
      console.log('✅ 已等待 Android 端停止')

      // 3. 重置状态
      isCameraOn.value = false
      streamStatus.value = 'inactive'
    } catch (error) {
      console.error('❌ 离开页面时停止摄像头失败:', error)
    } finally {
      cameraControlling.value = false
    }
  }

  // 4. 断开 LiveKit
  disconnectLiveKit()

  // 5. 关闭 WebSocket
  closeWebSocket()

  // 6. 清理 LiveKit 重连定时器
  if (liveKitReconnectTimer) {
    clearTimeout(liveKitReconnectTimer)
    liveKitReconnectTimer = null
  }
  clearCameraOnAckTimer()
  pendingCameraOnCommand.value = false

  console.log('✅ 组件卸载完成')
})

watch([videoHeight, videoRotationDeg], () => {
  if (videoPrefsLoaded.value) {
    hasVideoHeightPreference.value = true
  }
  saveVideoViewPreferences()
})
</script>

<style scoped lang="scss">
.device-detail {
  padding: 20px;

  // 响应式：小屏幕减少 padding
  @media (max-width: 768px) {
    padding: 12px;
  }

  @media (max-width: 480px) {
    padding: 8px;
  }
}

.camera-control {
  padding: 20px 0;

  .status-badge {
    margin-left: 20px;
  }

  // 响应式：移动端换行显示
  @media (max-width: 768px) {
    padding: 12px 0;

    .el-space {
      display: flex;
      flex-wrap: wrap;
      gap: 12px !important;
    }

    .status-badge {
      margin-left: 0;
      margin-top: 12px;
      width: 100%;
    }
  }

  @media (max-width: 480px) {
    .el-select,
    .el-button {
      flex: 1;
      min-width: 120px;
    }

    .el-button .el-icon {
      font-size: 12px;
    }
  }
}

.video-container {
  position: relative;
  width: 100%;
  background: #000;
  border-radius: 4px;
  overflow: hidden;
  transition: height 0.3s ease;

  // 纵向视频容器，支持动态高度调整

  // 响应式：桌面端默认高度
  @media (min-width: 1024px) {
    max-height: 85vh;
  }

  // 响应式：移动端优化
  @media (max-width: 1024px) {
    max-height: 80vh;
    border-radius: 0;
  }

  @media (max-width: 768px) {
    max-height: 75vh;
  }

  @media (max-width: 480px) {
    max-height: 70vh;
  }

  // 添加触摸支持
  touch-action: manipulation;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

.video-controls {
  margin: 6px 0 14px;
  align-items: center;
}

.video-element {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  background: #000;

  // 确保视频始终居中显示
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);

  // 纵向视频优化
  @media (orientation: portrait) {
    object-fit: contain;
  }

  // 横屏时也保持良好显示
  @media (orientation: landscape) {
    object-fit: contain;
  }
}

.video-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #999;
  background: #1a1a1a;

  .el-icon {
    margin-bottom: 12px;

    @media (max-width: 480px) {
      font-size: 36px !important;
    }
  }

  p {
    margin: 0;
    font-size: 14px;

    @media (max-width: 480px) {
      font-size: 12px;
    }
  }

  .el-button {
    margin-top: 16px;

    @media (max-width: 480px) {
      margin-top: 12px;
      font-size: 12px;
    }
  }
}

.bitrate-control {
  padding: 10px 0;

  > div {
    display: flex;
    align-items: center;
  }

  // 响应式：移动端换行显示
  @media (max-width: 768px) {
    .el-space {
      display: flex;
      flex-wrap: wrap;
      gap: 8px !important;
    }

    .el-input-number {
      min-width: 120px;
    }
  }

  @media (max-width: 480px) {
    padding: 8px 0;

    > div {
      flex-direction: column;
      align-items: flex-start;
    }

    .el-space {
      width: 100%;
      justify-content: space-between;
    }
  }
}

.bitrate-info {
  padding: 10px 0;

  // 响应式：移动端换行显示标签
  @media (max-width: 768px) {
    .el-space {
      display: flex;
      flex-wrap: wrap;
      gap: 8px !important;
    }
  }

  @media (max-width: 480px) {
    padding: 8px 0;

    .el-tag {
      font-size: 12px;
      padding: 4px 8px;
    }
  }
}

// 响应式：el-descriptions 在移动端适配
:deep(.el-descriptions) {
  @media (max-width: 768px) {
    .el-descriptions__label,
    .el-descriptions__content {
      font-size: 13px;
    }
  }

  @media (max-width: 480px) {
    .el-descriptions__label,
    .el-descriptions__content {
      font-size: 12px;
    }
  }
}

// 响应式：el-table 在移动端适配
:deep(.el-table) {
  @media (max-width: 768px) {
    font-size: 13px;
  }

  @media (max-width: 480px) {
    font-size: 12px;

    .el-table__cell {
      padding: 8px 0;
    }

    // 隐藏不重要的列
    .el-table__header-wrapper,
    .el-table__body-wrapper {
      .cell {
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
  }
}

// 响应式：el-card 在移动端适配
:deep(.el-card) {
  @media (max-width: 768px) {
    margin-bottom: 12px !important;
  }

  @media (max-width: 480px) {
    margin-bottom: 8px !important;

    .el-card__header {
      padding: 12px;
      font-size: 14px;

      // 为视频大小控制调整样式
      .el-slider {
        width: 100px !important;
      }
    }

    .el-card__body {
      padding: 12px;
    }
  }
}

// 响应式：el-row 在移动端适配
:deep(.el-row) {
  @media (max-width: 768px) {
    .el-col {
      margin-bottom: 16px;
    }
  }

  @media (max-width: 480px) {
    .el-col {
      margin-bottom: 12px;
    }
  }
}

// 响应式：el-page-header 在移动端适配
:deep(.el-page-header) {
  @media (max-width: 480px) {
    .el-page-header__left {
      margin-right: 8px;
    }

    .el-page-header__content {
      font-size: 16px;
    }
  }
}

// 响应式：el-slider 视频大小控制
:deep(.el-slider) {
  @media (max-width: 768px) {
    .el-slider__marks-text {
      font-size: 10px;
    }
  }

  @media (max-width: 480px) {
    .el-slider__button {
      width: 12px;
      height: 12px;
    }

    .el-slider__runway {
      height: 4px;
    }

    .el-slider__marks-text {
      font-size: 9px;
    }
  }
}

</style>

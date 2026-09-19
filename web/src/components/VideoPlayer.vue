<template>
  <div class="video-player">
    <div v-if="loading" class="loading-overlay">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>正在连接...</span>
    </div>
    
    <div v-if="error" class="error-overlay">
      <el-icon><Warning /></el-icon>
      <span>{{ error }}</span>
    </div>

    <!-- WebRTC 视频播放器 -->
    <video
      ref="videoPlayer"
      :muted="muted"
      :autoplay="autoplay"
      :controls="controls"
      @play="handlePlay"
      @pause="handlePause"
      class="video-element"
    />

    <!-- 播放控制栏 -->
    <div v-if="showControls" class="controls">
      <el-button-group>
        <el-button 
          :icon="isPlaying ? VideoPause : VideoPlay" 
          @click="togglePlay"
        />
        <el-button 
          :icon="isMuted ? MuteNotification : Notification" 
          @click="toggleMute"
        />
        <el-button 
          icon="FullScreen" 
          @click="toggleFullscreen"
        />
      </el-button-group>
    </div>

    <!-- 状态栏 -->
    <div v-if="showStatus" class="status-bar">
      <span>WebRTC 状态: {{ statusText }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  VideoPlay, VideoPause, 
  MuteNotification, Notification, 
  FullScreen, Loading, Warning 
} from '@element-plus/icons-vue'

// Props
interface Props {
  autoplay?: boolean
  muted?: boolean
  controls?: boolean
  showControls?: boolean
  showStatus?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  autoplay: true,
  muted: false,
  controls: true,
  showControls: true,
  showStatus: true
})

// Emits
const emit = defineEmits<{
  play: []
  pause: []
  error: [error: string]
  connected: []
  disconnected: []
}>()

// Refs
const videoPlayer = ref<HTMLVideoElement>()
const loading = ref(false)
const error = ref('')
const isPlaying = ref(false)
const isMuted = ref(props.muted)
const webrtcState = ref<'idle' | 'connecting' | 'connected' | 'failed'>('idle')

// WebRTC 相关
let peerConnection: RTCPeerConnection | null = null

// 计算属性
const statusText = computed(() => {
  switch (webrtcState.value) {
    case 'connecting': return '连接中...'
    case 'connected': return '已连接'
    case 'failed': return '连接失败'
    default: return '未连接'
  }
})

// 事件处理
const handlePlay = () => {
  isPlaying.value = true
  emit('play')
}

const handlePause = () => {
  isPlaying.value = false
  emit('pause')
}

const togglePlay = () => {
  const video = videoPlayer.value
  if (!video) return

  if (isPlaying.value) {
    video.pause()
  } else {
    video.play()
  }
}

const toggleMute = () => {
  const video = videoPlayer.value
  if (!video) return

  isMuted.value = !isMuted.value
  video.muted = isMuted.value
}

const toggleFullscreen = () => {
  const video = videoPlayer.value
  if (!video) return

  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    video.requestFullscreen()
  }
}

// WebRTC 连接逻辑
const connectWebRTC = async (deviceId: string, offer: string): Promise<void> => {
  try {
    webrtcState.value = 'connecting'
    loading.value = true
    error.value = ''

    // 创建 RTCPeerConnection
    peerConnection = new RTCPeerConnection({
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' }
      ]
    })

    // 处理 incoming stream
    peerConnection.ontrack = (event) => {
      const video = videoPlayer.value
      if (video && event.streams && event.streams[0]) {
        video.srcObject = event.streams[0]
        loading.value = false
        webrtcState.value = 'connected'
        emit('connected')
      }
    }

    // 设置远程描述
    const remoteDesc = new RTCSessionDescription({
      type: 'answer',
      sdp: offer
    })
    await peerConnection.setRemoteDescription(remoteDesc)

  } catch (err) {
    loading.value = false
    error.value = `WebRTC 连接失败: ${err}`
    webrtcState.value = 'failed'
    emit('error', error.value)
    ElMessage.error(error.value)
  }
}

// 暴露方法给父组件
defineExpose({
  connectWebRTC
})

// 组件卸载
onBeforeUnmount(() => {
  if (peerConnection) {
    peerConnection.close()
    peerConnection = null
    emit('disconnected')
  }
})
</script>

<style scoped lang="scss">
.video-player {
  position: relative;
  width: 100%;
  height: 100%;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.video-element {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.controls {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 12px;
  background: linear-gradient(transparent, rgba(0, 0, 0, 0.7));
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.3s;
  
  &:hover {
    opacity: 1;
  }
}

.status-bar {
  position: absolute;
  top: 12px;
  left: 12px;
  padding: 4px 12px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  font-size: 12px;
  border-radius: 4px;
}

.loading-overlay,
.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: white;
  font-size: 14px;
  z-index: 10;
}

.loading-overlay {
  background: rgba(0, 0, 0, 0.5);
}

.error-overlay {
  background: rgba(255, 0, 0, 0.2);
}

.is-loading {
  font-size: 32px;
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

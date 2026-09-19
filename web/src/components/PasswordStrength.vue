<template>
  <div class="password-strength">
    <div class="strength-bar">
      <div 
        class="strength-bar-fill" 
        :class="strengthClass"
        :style="{ width: strengthPercentage + '%' }"
      />
    </div>
    <div class="strength-text" :class="strengthClass">
      {{ strengthText }}
    </div>
    
    <!-- 密码要求提示 -->
    <div class="requirements">
      <div 
        v-for="req in requirements" 
        :key="req.key"
        class="requirement-item"
        :class="{ 'satisfied': req.satisfied }"
      >
        <el-icon>
          <component :is="req.satisfied ? CircleCheck : CircleClose" />
        </el-icon>
        <span>{{ req.text }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { CircleCheck, CircleClose } from '@element-plus/icons-vue'

interface Props {
  password: string
}

const props = defineProps<Props>()

// 密码强度等级
enum StrengthLevel {
  Weak = 0,
  Fair = 1,
  Good = 2,
  Strong = 3,
  VeryStrong = 4,
}

// 密码要求
const requirements = computed(() => [
  {
    key: 'length',
    text: '至少 8 个字符',
    satisfied: props.password.length >= 8,
  },
  {
    key: 'lowercase',
    text: '包含小写字母',
    satisfied: /[a-z]/.test(props.password),
  },
  {
    key: 'uppercase',
    text: '包含大写字母',
    satisfied: /[A-Z]/.test(props.password),
  },
  {
    key: 'number',
    text: '包含数字',
    satisfied: /[0-9]/.test(props.password),
  },
  {
    key: 'special',
    text: '包含特殊字符 (!@#$%^&*)',
    satisfied: /[!@#$%^&*]/.test(props.password),
  },
])

// 计算密码强度
const strength = computed(() => {
  let score = 0
  
  // 长度加分
  if (props.password.length >= 8) score++
  if (props.password.length >= 12) score++
  
  // 字符类型加分
  const hasLower = /[a-z]/.test(props.password)
  const hasUpper = /[A-Z]/.test(props.password)
  const hasNumber = /[0-9]/.test(props.password)
  const hasSpecial = /[!@#$%^&*]/.test(props.password)
  
  const charTypes = [hasLower, hasUpper, hasNumber, hasSpecial].filter(Boolean).length
  score += Math.min(charTypes - 1, 2) // 最多加2分
  
  return Math.min(score, StrengthLevel.VeryStrong) as StrengthLevel
})

// 强度百分比
const strengthPercentage = computed(() => {
  return ((strength.value + 1) / 5) * 100
})

// 强度样式类
const strengthClass = computed(() => {
  const classes = {
    [StrengthLevel.Weak]: 'weak',
    [StrengthLevel.Fair]: 'fair',
    [StrengthLevel.Good]: 'good',
    [StrengthLevel.Strong]: 'strong',
    [StrengthLevel.VeryStrong]: 'very-strong',
  }
  return classes[strength.value] || 'weak'
})

// 强度文字
const strengthText = computed(() => {
  const texts = {
    [StrengthLevel.Weak]: '弱',
    [StrengthLevel.Fair]: '中等',
    [StrengthLevel.Good]: '良好',
    [StrengthLevel.Strong]: '强',
    [StrengthLevel.VeryStrong]: '非常强',
  }
  return texts[strength.value] || '弱'
})
</script>

<style lang="scss" scoped>
.password-strength {
  margin-top: 8px;
}

.strength-bar {
  height: 4px;
  background: #e4e7ed;
  border-radius: 2px;
  overflow: hidden;
  margin-bottom: 8px;
}

.strength-bar-fill {
  height: 100%;
  transition: all 0.3s ease;
  
  &.weak {
    background: #f56c6c;
  }
  
  &.fair {
    background: #e6a23c;
  }
  
  &.good {
    background: #409eff;
  }
  
  &.strong {
    background: #67c23a;
  }
  
  &.very-strong {
    background: #67c23a;
  }
}

.strength-text {
  font-size: 12px;
  font-weight: 500;
  margin-bottom: 12px;
  
  &.weak {
    color: #f56c6c;
  }
  
  &.fair {
    color: #e6a23c;
  }
  
  &.good {
    color: #409eff;
  }
  
  &.strong,
  &.very-strong {
    color: #67c23a;
  }
}

.requirements {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 12px;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 6px;
}

.requirement-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #909399;
  transition: all 0.3s;
  
  .el-icon {
    font-size: 14px;
  }
  
  &.satisfied {
    color: #67c23a;
  }
}
</style>

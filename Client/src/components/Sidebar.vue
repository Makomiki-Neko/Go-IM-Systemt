<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getAvatarUrl } from '@/utils/avatar'

defineProps<{
  activePanel: 'chat' | 'contact' | 'notify'
  connected: boolean
  notifyCount: number
  chatUnreadNotify: boolean
}>()

const emit = defineEmits<{
  'update:activePanel': [panel: 'chat' | 'contact' | 'notify']
}>()

const router = useRouter()
const auth = useAuthStore()

function goProfile() {
  emit('update:activePanel', 'contact')
  window.dispatchEvent(new CustomEvent('im:showProfile'))
}

function logout() {
  auth.logout()
  router.replace('/login')
}
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <div class="user-avatar" @click="goProfile" :title="auth.userInfo?.name || '个人信息'">
        <img v-if="getAvatarUrl(auth.userInfo?.photo)" :src="getAvatarUrl(auth.userInfo?.photo)" class="avatar-img" />
        <span v-else>{{ auth.userInfo?.name?.[0] || auth.account?.[0] || 'U' }}</span>
      </div>
      <div :class="['status-dot', connected ? 'online' : 'offline']" :title="connected ? '已连接' : '未连接'" />
    </div>
    <div class="sidebar-tabs">
      <div
        :class="['tab', activePanel === 'chat' ? 'active' : '']"
        title="会话列表"
        @click="emit('update:activePanel', 'chat')"
      >
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
        </svg>
        <span v-if="chatUnreadNotify" class="chat-unread-dot" />
      </div>
      <div
        :class="['tab', activePanel === 'contact' ? 'active' : '']"
        title="好友列表"
        @click="emit('update:activePanel', 'contact')"
      >
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
          <circle cx="9" cy="7" r="4"/>
          <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
          <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
        </svg>
      </div>
      <div
        :class="['tab', activePanel === 'notify' ? 'active' : '']"
        title="通知"
        @click="emit('update:activePanel', 'notify')"
      >
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
          <path d="M13.73 21a2 2 0 0 1-3.46 0"/>
        </svg>
        <span v-if="notifyCount > 0" class="notify-badge">{{ notifyCount > 99 ? '99+' : notifyCount }}</span>
      </div>
    </div>
    <div class="sidebar-footer">
      <div class="tab" title="退出登录" @click="logout">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
          <polyline points="16 17 21 12 16 7"/>
          <line x1="21" y1="12" x2="9" y2="12"/>
        </svg>
      </div>
    </div>
  </div>
</template>

<style scoped>
.notify-badge, .chat-unread-dot {
  position: absolute;
}
.notify-badge {
  top: 2px;
  right: 4px;
  background: #e74c3c;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  min-width: 18px;
  height: 18px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}
.chat-unread-dot {
  top: 3px;
  right: 5px;
  width: 8px;
  height: 8px;
  background: #e74c3c;
  border-radius: 50%;
}
</style>

<script setup lang="ts">
import type { Conversation } from '@/types'
import { getAvatarUrl } from '@/utils/avatar'

defineProps<{
  conversations: Conversation[]
  currentId: string | null
}>()

const emit = defineEmits<{
  select: [userId: string]
}>()

function formatTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(Number(ts) * 1000)
  const now = new Date()
  if (d.toDateString() === now.toDateString()) return d.toLocaleTimeString().slice(0, 5)
  return `${d.getMonth() + 1}/${d.getDate()}`
}
</script>

<template>
  <div class="chat-list-panel">
    <div class="panel-header"><h3>会话列表</h3></div>
    <div class="chat-list">
      <div v-if="conversations.length === 0" class="empty-tip">暂无对话</div>
      <div
        v-for="conv in conversations" :key="conv.userId"
        :class="['chat-item', currentId === conv.userId ? 'active' : '']"
        @click="emit('select', conv.userId)"
      >
        <div class="chat-avatar">
          <img v-if="getAvatarUrl(conv.avatar)" :src="getAvatarUrl(conv.avatar)" class="avatar-img" />
          <span v-else>{{ conv.remark?.[0] || '?' }}</span>
          <span :class="['online-dot', conv.isOnline ? 'online' : '']" />
        </div>
        <div class="chat-info">
          <div class="chat-name-row">
            <span class="chat-name">{{ conv.remark }}</span>
            <span class="chat-time">{{ formatTime(conv.lastTime) }}</span>
          </div>
          <div class="chat-last-msg">{{ conv.lastMsg || '' }}</div>
        </div>
        <div v-if="conv.unread > 0" class="unread-badge">
          {{ conv.unread > 99 ? '99+' : conv.unread }}
        </div>
      </div>
    </div>
  </div>
</template>

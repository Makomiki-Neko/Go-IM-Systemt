<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { useFriendStore } from '@/stores/friend'
import { wsClient } from '@/api/ws'
import { getAvatarUrl } from '@/utils/avatar'
import type { ChatMessage } from '@/types'

const props = defineProps<{ userId: string }>()
const emit = defineEmits<{ close: [] }>()

const auth = useAuthStore()
const chatStore = useChatStore()
const friendStore = useFriendStore()

const text = ref('')
const bottomRef = ref<HTMLElement | null>(null)
const msgListRef = ref<HTMLElement | null>(null)
const loadingHistory = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)

const msgs = computed(() => chatStore.getMessages(props.userId))
const friend = computed(() => friendStore.friendList.find((f) => f.id === props.userId))

/* 打开时标记 */
onMounted(() => {
  chatStore.openChat(props.userId)
  scrollBottom()
})
onUnmounted(() => {
  chatStore.closeChat()
})

async function scrollBottom() {
  await nextTick()
  bottomRef.value?.scrollIntoView({ behavior: 'smooth' })
}
watch(() => msgs.value.length, () => scrollBottom())
watch(() => props.userId, () => {
  chatStore.openChat(props.userId)
  scrollBottom()
})

/* 发送文本消息 */
async function sendMessage() {
  const content = text.value.trim()
  if (!content || !auth.uid) return
  const tempId = Date.now() + Math.floor(Math.random() * 1000)
  const now = Math.floor(Date.now() / 1000)
  const optimisticMsg: ChatMessage = {
    msg_id: String(tempId),
    from_user_id: auth.uid,
    to_user_id: props.userId,
    msg_type: 1,
    content,
    send_time: now,
    status: 0,
  }
  chatStore.addOptimisticMsg(props.userId, tempId, optimisticMsg)
  text.value = ''
  chatStore.updateConversation(props.userId, { lastMsg: content, lastTime: now })

  const clientMsgId = tempId
  const sentReqId = wsClient.sendOnly('chat.SendPrivateMsg', {
    from_user_id: auth.uid,
    to_user_id: props.userId,
    msg_type: 1,
    content,
    client_msg_id: clientMsgId,
  })
  if (sentReqId === null) {
    chatStore.updateOptimisticMsg(props.userId, tempId, { _sending: false, _failed: true })
  } else {
    chatStore.trackPendingSend(clientMsgId, tempId, props.userId, sentReqId)
    /* 自动拉取新消息以获取服务器确认 */
    const conv = chatStore.conversations.find((c) => c.userId === props.userId)
    wsClient.sendOnly('chat.GetNewPrivateMsg', {
      user_id: auth.uid, from_user_id: props.userId,
      start_msg_id: conv?.lastMsgId || 0, limit: 50,
    })
  }
}

/* 拉取历史消息（滚动到顶部） */
async function loadHistory() {
  if (loadingHistory.value) return
  const list = chatStore.getMessages(props.userId)
  if (list.length === 0) return
  loadingHistory.value = true
  try {
    wsClient.sendOnly('chat.GetHistoryPrivateMsg', {
      user_id: auth.uid,
      from_user_id: props.userId,
      start_msg_id: list[0].msg_id,
      limit: 20,
    })
  } finally {
    loadingHistory.value = false
  }
}

/* 文件上传 */
async function uploadFile(file: File) {
  if (!auth.uid) return
  const fileId = Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
  const fileType = file.type.startsWith('image') ? 'Picture'
    : file.type.startsWith('video') ? 'Video'
    : file.type.startsWith('audio') ? 'Audio'
    : 'File'

  const result = await wsClient.send(`updateFile.${fileType}`, {
    file_id: fileId, file_name: file.name,
    file_size: Math.ceil(file.size / 1024), file_type: fileType,
  })
  const data = typeof result.data === 'string' ? JSON.parse(result.data) : result.data
  if (!data?.url) return

  try { await fetch(data.url, { method: 'PUT', body: file }) } catch { return }

  const tempId = Date.now() + Math.floor(Math.random() * 1000)
  const now = Math.floor(Date.now() / 1000)
  const msgType: number = fileType === 'Picture' ? 2 : fileType === 'Video' ? 4 : fileType === 'Audio' ? 3 : 5
  const optimisticMsg: ChatMessage = {
    msg_id: String(tempId), from_user_id: auth.uid, to_user_id: props.userId,
    msg_type: msgType, content: data.fileId, send_time: now, status: 0,
  }
  chatStore.addOptimisticMsg(props.userId, tempId, optimisticMsg)
  const sentReqId = wsClient.sendOnly('chat.SendPrivateMsg', {
    from_user_id: auth.uid, to_user_id: props.userId,
    msg_type: msgType, content: data.fileId, client_msg_id: tempId,
  })
  if (sentReqId === null) {
    chatStore.updateOptimisticMsg(props.userId, tempId, { _sending: false, _failed: true })
  } else {
    chatStore.trackPendingSend(tempId, tempId, props.userId, sentReqId)
    const conv = chatStore.conversations.find((c) => c.userId === props.userId)
    wsClient.sendOnly('chat.GetNewPrivateMsg', {
      user_id: auth.uid, from_user_id: props.userId,
      start_msg_id: conv?.lastMsgId || 0, limit: 50,
    })
  }
}

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (file) uploadFile(file)
  if (fileInput.value) fileInput.value.value = ''
}

const msgTypeMap: Record<number, string> = { 1: '文本', 2: '图片', 3: '语音', 4: '视频', 5: '文件' }

const zoomedImage = ref('')

function zoomImage(url: string) { zoomedImage.value = url }
function closeZoom() { zoomedImage.value = '' }
</script>

<template>
  <div class="message-area">
    <div class="message-header">
      <button class="btn-back" @click="emit('close')">‹</button>
      <div class="header-avatar">
        <img v-if="getAvatarUrl(friend?.avatar)" :src="getAvatarUrl(friend?.avatar)" class="avatar-img" />
        <span v-else>{{ friend?.remark?.[0] || friend?.friend_name?.[0] || '?' }}</span>
      </div>
      <div class="header-info">
        <h3>{{ friend?.remark || friend?.friend_name || '用户' }}</h3>
        <span :class="['online-status', friend?.online_statu ? 'online' : '']">
          {{ friend?.online_statu ? '在线' : '离线' }}
        </span>
      </div>
    </div>

    <div class="message-list" ref="msgListRef">
      <div v-if="loadingHistory" class="history-trigger loading">⏳ 加载中...</div>
      <div v-else-if="chatStore.noMoreHistorySet.has(props.userId) && msgs.length > 0" class="history-trigger exhausted">—— 无更多消息 ——</div>
      <div v-else-if="msgs.length > 0" class="history-trigger" @click="loadHistory">查看更多历史信息</div>
      <div v-if="msgs.length === 0" class="empty-tip">暂无消息，发送一条消息开始对话</div>
      <div
        v-for="msg in msgs" :key="msg.msg_id"
        :class="['msg-row', String(msg.from_user_id) === auth.uid ? 'self' : 'other']"
      >
        <div class="msg-avatar-col">
          <img v-if="String(msg.from_user_id) !== auth.uid && getAvatarUrl(friend?.avatar)" :src="getAvatarUrl(friend?.avatar)" class="msg-avatar-img" />
          <span v-else-if="String(msg.from_user_id) !== auth.uid">{{ friend?.remark?.[0] || friend?.friend_name?.[0] || '?' }}</span>
          <img v-else-if="getAvatarUrl(auth.userInfo?.photo)" :src="getAvatarUrl(auth.userInfo?.photo)" class="msg-avatar-img" />
          <span v-else class="self-label">我</span>
        </div>
        <div class="msg-body">
          <div class="msg-bubble">
            <div class="msg-content">
              <template v-if="msg.msg_type === 1">{{ msg.content }}</template>
              <template v-else-if="msg.msg_type === 2">
                <img :src="getAvatarUrl(msg.content) || msg.content" class="msg-image" alt="图片" @click.stop="zoomImage(getAvatarUrl(msg.content) || msg.content)" />
              </template>
              <template v-else>
                <span class="msg-file">[{{ msgTypeMap[msg.msg_type] || '文件' }}] {{ msg.content }}</span>
              </template>
            </div>
          </div>
          <div class="msg-time">
            {{ new Date(Number(msg.send_time) * 1000).toLocaleTimeString() }}
            <span v-if="msg._failed" class="msg-status msg-failed">·发送失败</span>
            <span v-else-if="msg._sending" class="msg-status msg-sending">·发送中</span>
            <span v-else-if="String(msg.from_user_id) === auth.uid" class="msg-status">
              {{ msg.status === 0 ? '·已发送' : msg.status === 1 ? '·已送达' : '·已读' }}
            </span>
          </div>
        </div>
      </div>
      <div ref="bottomRef" />
    </div>

    <div class="message-input-area">
      <button class="btn-file" @click="fileInput?.click()">📎</button>
      <input ref="fileInput" type="file" style="display:none" @change="onFileChange" />
      <input v-model="text" class="msg-input" placeholder="输入消息..." @keydown.enter="sendMessage" />
      <button class="btn-send" @click="sendMessage">发送</button>
    </div>

    <!-- 图片放大 -->
    <div v-if="zoomedImage" class="zoom-overlay" @click.self="closeZoom">
      <button class="zoom-close" @click="closeZoom">✕</button>
      <img :src="zoomedImage" class="zoom-image" @click.self="closeZoom" />
    </div>
  </div>
</template>

<style scoped>
.btn-back { background: none; border: none; font-size: 22px; cursor: pointer; color: #666; padding: 0 4px; }
.btn-back:hover { color: #333; }
.message-header { display: flex; align-items: center; gap: 10px; padding: 12px 16px; border-bottom: 1px solid #e8e8e8; background: #fff; }
.header-avatar { width: 36px; height: 36px; border-radius: 8px; background: #667eea; color: #fff; display: flex; align-items: center; justify-content: center; font-size: 14px; font-weight: 600; overflow: hidden; flex-shrink: 0; }
.header-avatar .avatar-img { width: 100%; height: 100%; object-fit: cover; }
.header-info { flex: 1; }
.header-info h3 { font-size: 15px; margin: 0; }
.header-info .online-status { font-size: 11px; color: #ccc; }
.header-info .online-status.online { color: #2ecc71; }
.msg-image { max-width: 100%; max-height: 300px; border-radius: 8px; cursor: pointer; display: block; }
.msg-image:hover { opacity: 0.9; }
.btn-file { background: none; border: none; font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; }
.btn-file:hover { background: #f0f0f0; }
.msg-sending { color: #999; }
.msg-failed { color: #e74c3c; }
.msg-avatar-col { flex-shrink: 0; width: 36px; height: 36px; border-radius: 8px; overflow: hidden; display: flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; background: #667eea; color: #fff; }
.msg-avatar-img { width: 100%; height: 100%; object-fit: cover; }
.self-label { font-size: 12px; }
.zoom-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.85); display: flex; align-items: center; justify-content: center; z-index: 2000; cursor: zoom-out; }
.zoom-close { position: absolute; top: 16px; right: 20px; background: none; border: none; color: #fff; font-size: 28px; cursor: pointer; z-index: 2001; }
.zoom-image { max-width: 90vw; max-height: 90vh; object-fit: contain; border-radius: 4px; }
.history-trigger { text-align: center; padding: 8px 0; font-size: 12px; cursor: pointer; user-select: none; }
.history-trigger.loading { color: #999; cursor: default; }
.history-trigger.exhausted { color: #bbb; cursor: default; }
.history-trigger:not(.loading):not(.exhausted) { color: #667eea; }
.history-trigger:not(.loading):not(.exhausted):hover { color: #5a6fd6; text-decoration: underline; }
</style>

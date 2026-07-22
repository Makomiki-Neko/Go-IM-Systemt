<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import type { Conversation, ChatMessage, WsPushMessage, FriendInfo } from '@/types'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'
import { useFriendStore } from '@/stores/friend'
import { wsClient } from '@/api/ws'
import { getUserInfoApi } from '@/api/user'
import { getUnreadPrivateMsgNumberApi } from '@/api/chat'
import { getFriendListApi, getFriendApplyListApi } from '@/api/friend'
import { heartApi } from '@/api/auth'
import Sidebar from '@/components/Sidebar.vue'
import ChatListPanel from '@/components/ChatListPanel.vue'
import MessageArea from '@/components/MessageArea.vue'
import ContactPanel from '@/components/ContactPanel.vue'
import FriendDetail from '@/components/FriendDetail.vue'
import UserProfile from '@/components/UserProfile.vue'
import NotifyPanel from '@/components/NotifyPanel.vue'

const router = useRouter()
const auth = useAuthStore()
const chat = useChatStore()
const friend = useFriendStore()

const middlePanel = ref<'chat' | 'contact' | 'profile' | 'notify'>('chat')
const rightType = ref<'chat' | 'friend' | null>(null)
const rightChatUserId = ref<string | null>(null)
const rightFriend = ref<FriendInfo | null>(null)
const wsConnected = ref(false)
const chatUnreadNotify = ref(false)
let heartTimer: ReturnType<typeof setInterval> | null = null

/* ======== 初始化 ======== */
onMounted(async () => {
  if (!auth.accessToken) { router.replace('/login'); return }

  wsClient.onConnected = () => {
    wsConnected.value = true
    friend.resetPulled()
  }
  wsClient.onDisconnected = () => { wsConnected.value = false }
  wsClient.connect()

  wsClient.on('chat.privateMsg', handlePrivateMsg)
  wsClient.on('chat.privateUnreceiveMsgBlock', handleNewMsgBlock)
  wsClient.on('chat.privateHistoryMsgBlock', handleHistoryMsgBlock)
  wsClient.on('event.friendApply', handleFriendApply)
  wsClient.on('ack.Msg', handleMsgAck)
  wsClient.on('file.UpdateUrl', handleFileUrl)
  wsClient.on('Error', handleError)

  window.addEventListener('im:showProfile', showProfile)
  window.addEventListener('im:ack', handleAckEvent)
  window.addEventListener('im:showNotify', goToNotify)

  await loadInitialData()
  heartTimer = setInterval(refreshToken, 180000)
})

onUnmounted(() => {
  wsClient.disconnect()
  chat.stopAckTimer()
  if (heartTimer) { clearInterval(heartTimer); heartTimer = null }
  window.removeEventListener('im:showProfile', showProfile)
  window.removeEventListener('im:ack', handleAckEvent)
  window.removeEventListener('im:showNotify', goToNotify)
})

/* ======== 数据加载 ======== */
async function loadInitialData() {
  try {
    const [friendRes, unreadRes, infoRes] = await Promise.all([
      getFriendListApi(1, 200),
      getUnreadPrivateMsgNumberApi(),
      getUserInfoApi(auth.account),
    ])
    if (infoRes.data) auth.setUserInfo(infoRes.data)
    if (friendRes.data?.list) {
      friend.setFriendList(friendRes.data.list)
      const convs: Conversation[] = friendRes.data.list.map((f) => ({
        userId: f.id,
        friendName: f.friend_name,
        remark: f.remark || f.friend_name,
        avatar: f.avatar,
        lastMsg: '',
        lastTime: 0,
        unread: 0,
        isOnline: f.online_statu,
        lastMsgId: '0',
        pulledThisSession: false,
      }))
      chat.setConversations(convs)
    }
    try {
      const applyRes = await getFriendApplyListApi(1, 200)
      if (applyRes.data?.list) friend.setApplyList(applyRes.data.list)
    } catch {}
    if (unreadRes.data?.list) {
      for (const u of unreadRes.data.list) {
        const fid = String(u.id)
        chat.setUnread(fid, u.count)
        chat.updateConversation(fid, { unread: u.count })
        if (u.count > 0) {
          const conv = chat.conversations.find((c) => c.userId === fid)
          const startMsgId = conv?.lastMsgId || '0'
          wsClient.sendOnly('chat.GetNewPrivateMsg', {
            user_id: auth.uid, from_user_id: fid, start_msg_id: startMsgId, limit: 50,
          })
          friend.markPulled(fid)
        }
      }
    }
  } catch {}
}

/* ======== Token 刷新 ======== */
let heartRetries = 0
async function refreshToken() {
  try {
    const { data } = await heartApi({
      uid: auth.uid!, account: auth.account, refresh_token: auth.refreshToken, device_id: 'web', platform: 'web',
    })
    if (data.access_token) auth.setAccessToken(data.access_token)
    heartRetries = 0
  } catch {
    heartRetries++
    if (heartRetries < 3) {
      setTimeout(refreshToken, 30000)
    } else {
      auth.logout()
      router.replace('/login')
    }
  }
}

/* ======== WS 消息处理 ======== */
function ensureConversation(userId: string, data: any) {
  let conv = chat.conversations.find((c) => c.userId === userId)
  if (!conv) {
    conv = {
      userId,
      friendName: data.friend_name || '',
      remark: data.friend_name || '',
      avatar: data.avatar || '',
      lastMsg: '',
      lastTime: 0,
      unread: 0,
      isOnline: false,
      lastMsgId: '0',
      pulledThisSession: false,
    }
    chat.upsertConversation(conv)
  }
  return conv
}

function handlePrivateMsg(msg: WsPushMessage) {
  const data = Array.isArray(msg.data) ? msg.data[0] : typeof msg.data === 'string' ? JSON.parse(msg.data) : msg.data
  const fromId = String(data.from_user_id)
  const myUid = auth.uid!
  if (String(data.from_user_id) !== myUid) {
    ensureConversation(fromId, data)
    chat.pushMessage(fromId, data as ChatMessage)
    chat.updateConversation(fromId, { lastMsg: data.content, lastTime: data.send_time, lastMsgId: String(data.msg_id) })
    if (chat.currentChatId !== fromId) {
      const cur = chat.unreadMap.get(fromId) || 0
      chat.setUnread(fromId, cur + 1)
      chat.updateConversation(fromId, { unread: cur + 1 })
    }
    if (middlePanel.value !== 'chat') {
      chatUnreadNotify.value = true
    }
  }
}
function handleNewMsgBlock(msg: WsPushMessage) {
  const data = typeof msg.data === 'string' ? JSON.parse(msg.data) : msg.data
  const list = Array.isArray(data) ? data : [data]
  if (!list.length) return
  const myUid = auth.uid!
  const first = list[0] as ChatMessage
  const partnerId = String(first.from_user_id) === myUid ? String(first.to_user_id) : String(first.from_user_id)
  for (const m of list) {
    const cm = m as ChatMessage
    if (String(cm.from_user_id) === myUid) {
      chat.confirmOptimisticMsg(partnerId, cm)
    }
  }
  const existing = chat.getMessages(partnerId)
  const existingIds = new Set(existing.map((m) => String(m.msg_id)))
  const newMsgs = (list as ChatMessage[]).filter((m) => !existingIds.has(String(m.msg_id)))
  if (newMsgs.length > 0) {
    chat.appendMessages(partnerId, newMsgs)
  }
  const last = list[list.length - 1] as ChatMessage
  chat.updateConversation(partnerId, { lastMsg: last.content, lastTime: last.send_time, lastMsgId: String(last.msg_id) })
}
function handleHistoryMsgBlock(msg: WsPushMessage) {
  const data = typeof msg.data === 'string' ? JSON.parse(msg.data) : msg.data
  const list = Array.isArray(data) ? data : [data]
  if (!list.length) {
    chat.noMoreHistorySet.add(chat.currentChatId!)
    return
  }
  chat.prependMessages(String(list[0].from_user_id), list)
}
function handleFriendApply() {
  getFriendApplyListApi(1, 200).then((res) => {
    if (res.data?.list) friend.setApplyList(res.data.list)
  })
}
function handleMsgAck(msg: any) {
  const data = typeof msg.data === 'string' ? JSON.parse(msg.data) : msg.data
  const clientMsgId = Number(data.client_msg_id)
  if (clientMsgId) {
    chat.confirmByClientMsgId(clientMsgId, data.msg_id, Number(data.send_time))
  }
}
function handleFileUrl(msg: WsPushMessage) {
  const data = typeof msg.data === 'string' ? JSON.parse(msg.data) : msg.data
  window.dispatchEvent(new CustomEvent('im:fileUrl', { detail: data }))
}
function handleError(msg: any) {
  console.warn('WS error:', msg)
  if (msg.req_id) {
    chat.markFailedByReqId(msg.req_id)
  }
}
function handleAckEvent(e: Event) {
  const { target_id, msg_id } = (e as CustomEvent).detail
  wsClient.sendOnly('ack.PrivateMsgRead', { target_id, msg_id })
}

/* ======== 面板切换（带动态刷新） ======== */
function onSelectChat(userId: string) {
  chat.openChat(userId)
  rightType.value = 'chat'
  rightChatUserId.value = userId
  rightFriend.value = null
  if (!friend.hasPulled(userId)) {
    const conv = chat.conversations.find((c) => c.userId === userId)
    const startMsgId = conv?.lastMsgId || '0'
    wsClient.sendOnly('chat.GetNewPrivateMsg', {
      user_id: auth.uid, from_user_id: userId, start_msg_id: startMsgId, limit: 50,
    })
    friend.markPulled(userId)
  }
}

function showFriendDetail(f: FriendInfo) {
  rightType.value = 'friend'
  rightFriend.value = f
  rightChatUserId.value = null
  chat.closeChat()
}

function startChatFromFriend(userId: string) {
  const f = friend.friendList.find((x) => x.id === userId)
  if (f) {
    chat.upsertConversation({
      userId: f.id, friendName: f.friend_name, remark: f.remark || f.friend_name,
      avatar: f.avatar, lastMsg: '', lastTime: 0, unread: 0, isOnline: f.online_statu,
      lastMsgId: '0', pulledThisSession: false,
    })
  }
  middlePanel.value = 'chat'
  onSelectChat(userId)
}

function showProfile() {
  middlePanel.value = 'profile'
  rightType.value = null
  rightChatUserId.value = null
  rightFriend.value = null
  chat.closeChat()
}

function goToNotify() {
  onMiddlePanelChange('notify')
}

function closeRight() {
  rightType.value = null
  rightChatUserId.value = null
  rightFriend.value = null
}

async function onMiddlePanelChange(panel: 'chat' | 'contact' | 'notify') {
  middlePanel.value = panel
  if (panel === 'chat') {
    chatUnreadNotify.value = false
    if (rightType.value === 'friend') {
      closeRight()
    }
    // 刷新未读数量
    try {
      const { data } = await getUnreadPrivateMsgNumberApi()
      if (data?.list) {
        for (const u of data.list) {
          const fid = String(u.id)
          chat.setUnread(fid, u.count)
          chat.updateConversation(fid, { unread: u.count })
        }
      }
    } catch {}
  } else if (panel === 'contact') {
    closeRight()
    // 刷新好友列表
    try {
      const { data } = await getFriendListApi(1, 200)
      if (data?.list) friend.setFriendList(data.list)
    } catch {}
  } else if (panel === 'notify') {
    closeRight()
    // 刷新通知
    try {
      const { data } = await getFriendApplyListApi(1, 200)
      if (data?.list) friend.setApplyList(data.list)
    } catch {}
  }
}
</script>

<template>
  <div class="im-layout">
    <Sidebar
      :active-panel="middlePanel === 'profile' ? 'contact' : middlePanel"
      :connected="wsConnected"
      :notify-count="friend.applyList.length"
      :chat-unread-notify="chatUnreadNotify"
      @update:active-panel="onMiddlePanelChange"
    />

    <div class="main-panel">
      <ChatListPanel
        v-if="middlePanel === 'chat'"
        :conversations="chat.sortedConversations"
        :current-id="rightChatUserId"
        @select="onSelectChat"
      />
      <ContactPanel
        v-else-if="middlePanel === 'contact'"
        @show-detail="showFriendDetail"
        @start-chat="startChatFromFriend"
      />
      <NotifyPanel
        v-else-if="middlePanel === 'notify'"
      />
      <UserProfile
        v-else-if="middlePanel === 'profile'"
      />
    </div>

    <div v-if="rightType" class="message-panel">
      <MessageArea
        v-if="rightType === 'chat' && rightChatUserId"
        :user-id="rightChatUserId"
        @close="closeRight"
      />
      <FriendDetail
        v-else-if="rightType === 'friend' && rightFriend"
        :friend="rightFriend"
        @start-chat="startChatFromFriend"
        @close="closeRight"
      />
    </div>
    <div v-else class="message-panel message-panel-empty">
      <div class="no-chat-selected">
        <div class="empty-icon">&#x1F4AC;</div>
        <p v-if="middlePanel === 'chat'">选择一个对话开始聊天</p>
        <p v-else-if="middlePanel === 'contact'">选择一个好友查看信息</p>
        <p v-else-if="middlePanel === 'notify'">通知中心</p>
        <p v-else>个人信息</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-panel-empty { background: #f5f5f5; }
</style>

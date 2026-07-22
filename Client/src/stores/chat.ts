import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Conversation, ChatMessage } from '@/types'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const currentChatId = ref<string | null>(null)
  const messagesMap = ref<Map<string, ChatMessage[]>>(new Map())
  const unreadMap = ref<Map<string, number>>(new Map())
  const lastReadMsgId = ref<string>('0')

  /** 等待服务端 ACK 的发送消息 <client_msg_id, { tempId, userId, reqId, timeout }> */
  const pendingSends = ref<Map<number, { tempId: number; userId: string; reqId: number; timeout: ReturnType<typeof setTimeout> }>>(new Map())
  const noMoreHistorySet = ref<Set<string>>(new Set())

  let ackTimer: ReturnType<typeof setInterval> | null = null

  const currentMessages = computed(() => {
    if (!currentChatId.value) return []
    return messagesMap.value.get(currentChatId.value) || []
  })

  const currentUnread = computed(() => {
    if (!currentChatId.value) return 0
    return unreadMap.value.get(currentChatId.value) || 0
  })

  function setConversations(list: Conversation[]) {
    conversations.value = list
  }

  function updateConversation(userId: string, patch: Partial<Conversation>) {
    const idx = conversations.value.findIndex((c) => c.userId === userId)
    if (idx >= 0) {
      conversations.value[idx] = { ...conversations.value[idx], ...patch }
    }
  }

  function upsertConversation(conv: Conversation) {
    const idx = conversations.value.findIndex((c) => c.userId === conv.userId)
    if (idx >= 0) {
      conversations.value[idx] = conv
    } else {
      conversations.value.push(conv)
    }
  }

  function appendMessages(userId: string, msgs: ChatMessage[]) {
    const existing = messagesMap.value.get(userId) || []
    const existingIds = new Set(existing.map((m) => String(m.msg_id)))
    const newMsgs = msgs.filter((m) => !existingIds.has(String(m.msg_id)))
    if (newMsgs.length > 0) {
      messagesMap.value.set(userId, [...existing, ...newMsgs])
    }
  }

  function prependMessages(userId: string, msgs: ChatMessage[]) {
    const existing = messagesMap.value.get(userId) || []
    const existingIds = new Set(existing.map((m) => String(m.msg_id)))
    const newMsgs = msgs.filter((m) => !existingIds.has(String(m.msg_id))).reverse()
    if (newMsgs.length > 0) {
      messagesMap.value.set(userId, [...newMsgs, ...existing])
    }
  }

  function getMessages(userId: string): ChatMessage[] {
    return messagesMap.value.get(userId) || []
  }

  function pushMessage(userId: string, msg: ChatMessage) {
    const list = messagesMap.value.get(userId) || []
    if (list.some((m) => String(m.msg_id) === String(msg.msg_id))) return
    list.push(msg)
    messagesMap.value.set(userId, list)
  }

  function addOptimisticMsg(userId: string, tempId: number, msg: ChatMessage) {
    const list = messagesMap.value.get(userId) || []
    list.push({ ...msg, _sending: true })
    messagesMap.value.set(userId, list)
  }

  function updateOptimisticMsg(userId: string, tempId: number, patch: Partial<ChatMessage>) {
    const list = messagesMap.value.get(userId)
    if (!list) return
    const idx = list.findIndex((m) => Number(m.msg_id) === tempId)
    if (idx >= 0) {
      list[idx] = { ...list[idx], ...patch }
      messagesMap.value.set(userId, [...list])
    }
  }

  function confirmOptimisticMsg(userId: string, realMsg: ChatMessage) {
    const list = messagesMap.value.get(userId)
    if (!list) return false
    const idx = list.findIndex(
      (m) => m._sending && m.from_user_id === realMsg.from_user_id
        && m.to_user_id === realMsg.to_user_id
        && m.content === realMsg.content
        && Math.abs(m.send_time - realMsg.send_time) < 5,
    )
    if (idx >= 0) {
      list[idx] = { ...realMsg, _sending: false }
      messagesMap.value.set(userId, [...list])
      return true
    }
    return false
  }

  /** 通过 req_id 匹配错误响应，标记消息为发送失败 */
  function markFailedByReqId(reqId: number) {
    for (const [clientMsgId, pending] of pendingSends.value.entries()) {
      if (pending.reqId === reqId) {
        clearTimeout(pending.timeout)
        pendingSends.value.delete(clientMsgId)
        updateOptimisticMsg(pending.userId, pending.tempId, { _sending: false, _failed: true })
        return true
      }
    }
    return false
  }

  /** 通过 client_msg_id 确认消息发送成功 */
  function confirmByClientMsgId(clientMsgId: number, realMsgId: string, realSendTime: number) {
    const pending = pendingSends.value.get(clientMsgId)
    if (!pending) return
    clearTimeout(pending.timeout)
    pendingSends.value.delete(clientMsgId)

    const list = messagesMap.value.get(pending.userId)
    if (!list) return
    const idx = list.findIndex((m) => m._sending && Number(m.msg_id) === pending.tempId)
    if (idx >= 0) {
      list[idx] = {
        ...list[idx],
        msg_id: realMsgId,
        send_time: realSendTime,
        _sending: false,
        _failed: false,
      }
      messagesMap.value.set(pending.userId, [...list])
    }
  }

  /** 注册待确认的发送消息 */
  function trackPendingSend(clientMsgId: number, tempId: number, userId: string, reqId: number) {
    const timeout = setTimeout(() => {
      const p = pendingSends.value.get(clientMsgId)
      if (p) {
        pendingSends.value.delete(clientMsgId)
        updateOptimisticMsg(userId, tempId, { _sending: false, _failed: true })
      }
    }, 20000)
    pendingSends.value.set(clientMsgId, { tempId, userId, reqId, timeout })
  }

  function setUnread(userId: string, count: number) {
    unreadMap.value.set(userId, count)
  }
  function clearUnread(userId: string) {
    unreadMap.value.set(userId, 0)
    const c = conversations.value.find((x) => x.userId === userId)
    if (c) c.unread = 0
  }

  function openChat(userId: string) {
    currentChatId.value = userId
    lastReadMsgId.value = '0'
    clearUnread(userId)
    noMoreHistorySet.value.delete(userId)
    startAckTimer(userId)
  }

  function closeChat() {
    stopAckTimer()
    currentChatId.value = null
    lastReadMsgId.value = '0'
  }

  function startAckTimer(userId: string) {
    stopAckTimer()
    ackTimer = setInterval(() => {
      const msgs = messagesMap.value.get(userId)
      if (!msgs || msgs.length === 0) return
      const latest = msgs[msgs.length - 1]
      if (latest.msg_id !== lastReadMsgId.value && Number(latest.msg_id) > 0) {
        lastReadMsgId.value = latest.msg_id
        window.dispatchEvent(
          new CustomEvent('im:ack', { detail: { target_id: userId, msg_id: latest.msg_id } }),
        )
      }
    }, 5000)
  }

  function stopAckTimer() {
    if (ackTimer) { clearInterval(ackTimer); ackTimer = null }
  }

  const sortedConversations = computed(() => {
    return [...conversations.value].sort((a, b) => b.lastTime - a.lastTime)
  })

  return {
    conversations, currentChatId, currentMessages, currentUnread,
    unreadMap, lastReadMsgId, pendingSends,
    setConversations, updateConversation, upsertConversation,
    appendMessages, prependMessages, getMessages, pushMessage,
    addOptimisticMsg, updateOptimisticMsg, confirmOptimisticMsg,
    confirmByClientMsgId, trackPendingSend, markFailedByReqId,
    setUnread, clearUnread,
    openChat, closeChat,
    startAckTimer, stopAckTimer,
    sortedConversations, noMoreHistorySet,
  }
})

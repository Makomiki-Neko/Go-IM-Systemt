<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import Icon from './Icon.vue'
import Avatar from './Avatar.vue'
import ChatAttachment from './ChatAttachment.vue'
import { state, bucket, send, upload, history, markRead, run, sync, groupList, toast } from '../lib/store'
import { array, object, str, type Conversation, type Message } from '../lib/protocol'
const props = defineProps<{ conversation: Conversation }>()
defineEmits<{ details: [] }>()
const draftKey = computed(() => `qiyu.draft:${state.session?.uid}:${bucket(props.conversation)}`)
const draft = ref(sessionStorage.getItem(draftKey.value) || '')
const messages = computed(() => state.messages[bucket(props.conversation)] || [])
const searchOpen = ref(false), search = ref(''), sending = ref(false), loadingHistory = ref(false), atBottom = ref(true), picker = ref(false), mentionsOpen = ref(false)
const selectedMentions = ref<string[]>([]), members = ref<{ id: string; name: string }[]>([])
const scroll = ref<HTMLElement>(), input = ref<HTMLTextAreaElement>(), file = ref<HTMLInputElement>()
let readBusy = false
const visible = computed(() => messages.value.filter(m => !search.value || m.content.toLowerCase().includes(search.value.toLowerCase())))
const self = (m: Message) => m.sender === state.session?.uid
const name = (id: string) => id === 'ai' ? props.conversation.name : id === state.session?.uid ? state.profile.name || '我' : state.friends.find(f => f.id === id)?.name || `成员 ${id.slice(-5)}`
const avatar = (id: string) => id === 'ai' ? props.conversation.avatar : id === state.session?.uid ? state.profile.photo : state.friends.find(f => f.id === id)?.avatar
const time = (ts: number) => new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
const date = (ts: number) => new Date(ts).toLocaleDateString('zh-CN', { month: 'long', day: 'numeric' })
// Preserve paragraph breaks and first-line indentation; hide only leading blank lines.
const replyText = (m: Message) => m.sender === 'ai' ? m.content.replace(/^(?:[ \t]*\r?\n)+/, '') : m.content
const stateText = (m: Message) => m.sender==='ai' ? (m.state==='queued' ? '等待回复' : m.state==='running' ? (replyText(m).trim() ? '正在回复' : '思考中') : m.state==='failed' ? '回复失败' : m.state==='unknown' ? '等待确认' : '回复完成') : ({ sending: '发送中', processing: '处理中', queued: '等待回复', running: '正在思考', succeeded: '已发送', failed: '未发送成功', unknown: '结果待确认' }[m.state || 'succeeded'])
watch(draft, v => sessionStorage.setItem(draftKey.value, v))
watch(() => props.conversation.version, () => { draft.value = sessionStorage.getItem(draftKey.value) || ''; selectedMentions.value = [] })
async function bottom() { await nextTick(); if (scroll.value) scroll.value.scrollTop = scroll.value.scrollHeight; atBottom.value = true; await read() }
async function read() {
  if (readBusy || !atBottom.value || search.value || loadingHistory.value) return
  readBusy = true
  try { await run(() => markRead(props.conversation)) } finally { readBusy = false }
}
function onScroll() { const el = scroll.value; if (el) { atBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight < 70; if (atBottom.value) void read() } }
watch(() => [messages.value.length, messages.value.map(m=>`${m.content.length}:${m.reasoning?.length || 0}:${m.state}`).join('|'), state.syncing[bucket(props.conversation)]], () => { if (atBottom.value && !loadingHistory.value) void bottom() })
async function loadHistory() {
  loadingHistory.value = true
  const oldHeight = scroll.value?.scrollHeight || 0, oldTop = scroll.value?.scrollTop || 0
  await run(() => history(props.conversation)); await nextTick()
  if (scroll.value) scroll.value.scrollTop = oldTop + scroll.value.scrollHeight - oldHeight
  loadingHistory.value = false
}
async function submit() {
  if (!draft.value.trim() || sending.value) return
  sending.value = true
  const text = draft.value
  const ok = await run(() => send(props.conversation, text, 1, [...selectedMentions.value]))
  if (ok) { draft.value = ''; selectedMentions.value = []; picker.value = false; await bottom() }
  sending.value = false; input.value?.focus()
}
function key(e: KeyboardEvent) { if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229) { e.preventDefault(); void submit() } }
async function attachment(e: Event) {
  const target = e.target as HTMLInputElement, chosen = target.files?.[0]; target.value = ''
  if (!chosen) return
  sending.value = true; await run(() => upload(props.conversation, chosen)); sending.value = false; await bottom()
}
async function mentionPicker() {
  mentionsOpen.value = !mentionsOpen.value
  if (!mentionsOpen.value) return
  if (state.demo) { members.value = state.friends.map(f => ({ id: f.id, name: f.name })); return }
  await run(async () => { const rows = await groupList('members', props.conversation.id); members.value = rows.map(r => ({ id: str(r.user_id), name: str(r.group_nick) || name(str(r.user_id)) })) })
}
function mention(m: { id: string; name: string }) { if (!selectedMentions.value.includes(m.id)) selectedMentions.value.push(m.id); draft.value += `@${m.name} `; mentionsOpen.value = false; input.value?.focus() }
async function retry(m: Message) {
  if (m.state === 'failed') await run(() => send(props.conversation, m.content, m.type, m.mentions || []))
  else if (props.conversation.kind === 'private') { await run(() => sync(props.conversation)); toast('已同步记录。私聊结果仍不明确时，请确认后再手动重发。') }
  else await run(() => send(props.conversation, m.content, m.type, m.mentions || [], m))
}
function focusChanged() { if (document.visibilityState === 'visible') void read() }
onMounted(() => { void bottom(); document.addEventListener('visibilitychange', focusChanged); window.addEventListener('focus', focusChanged) })
onUnmounted(() => { document.removeEventListener('visibilitychange', focusChanged); window.removeEventListener('focus', focusChanged) })
</script>
<template><section class="chat-pane">
  <header class="chat-header"><button class="icon-button mobile-back" aria-label="返回会话列表" @click="state.active = ''"><Icon name="ArrowLeft"/></button><Avatar :name="conversation.name" :src="conversation.avatar" :kind="conversation.kind" size="sm"/><div class="chat-title"><h2>{{ conversation.name }}<span v-if="conversation.kind === 'group'" class="tag">群聊</span><span v-if="conversation.kind === 'ai'" class="tag">AI</span></h2><p><i v-if="conversation.kind === 'private' && conversation.online" class="tiny-online"/>{{ conversation.kind === 'group' ? '一起分享，让灵感发生' : conversation.kind === 'ai' ? '陪你整理思路，探索新的可能' : conversation.online ? '在线 · 随时聊聊' : '离线 · 消息将在上线后可见' }}</p></div><div class="header-actions"><button class="icon-button" aria-label="搜索当前聊天" :class="{ active: searchOpen }" @click="searchOpen = !searchOpen; search = ''"><Icon name="Search"/></button><span class="action-divider"/><button class="icon-button" aria-label="会话详情" @click="$emit('details')"><Icon name="MoreHorizontal" :size="23"/></button></div></header>
  <div v-if="searchOpen" class="message-search"><Icon name="Search" :size="17"/><input v-model="search" placeholder="搜索本设备已加载的消息" aria-label="搜索消息"/><span>{{ visible.length }} 条</span><button class="icon-button" aria-label="关闭消息搜索" @click="searchOpen = false; search = ''"><Icon name="X" :size="16"/></button></div>
  <div class="chat-body" ref="scroll" @scroll="onScroll">
    <div class="history-control"><button v-if="!state.noHistory[bucket(conversation)]" class="text-button" :disabled="loadingHistory || state.syncing[bucket(conversation)]" @click="loadHistory"><Icon :name="loadingHistory ? 'LoaderCircle' : 'RefreshCw'" :class="{ spin: loadingHistory }" :size="13"/>{{ loadingHistory ? '正在加载…' : '查看更早的消息' }}</button><span v-else>这段对话，从这里开始</span></div>
    <div v-if="!visible.length" class="chat-empty"><Icon :name="conversation.kind === 'ai' ? 'Sparkles' : 'Coffee'" :size="34"/><h3>{{ search ? '没有找到相关消息' : '给对话一个温暖的开场' }}</h3><p>{{ search ? '试试其他关键词，或加载更早的消息。' : conversation.kind === 'ai' ? '有什么想法想一起探索？' : '不必想太多，一句你好就很好。' }}</p><button v-if="!search" class="button secondary" @click="draft = conversation.kind === 'ai' ? '帮我梳理一下今天的计划。' : '你好，很高兴在这里遇见你。'; input?.focus()">{{ conversation.kind === 'ai' ? '一起梳理计划' : '打个招呼' }}<Icon name="ArrowUpRight" :size="15"/></button></div>
    <template v-for="(m, i) in visible" :key="m.id">
      <div v-if="i === 0 || date(m.time) !== date(visible[i-1].time)" class="date-separator"><span>{{ date(m.time) }}</span></div>
      <article class="message-row" :class="{ own: self(m) }"><Avatar :name="name(m.sender)" :src="avatar(m.sender)" :kind="m.sender === 'ai' ? 'ai' : 'private'" size="sm"/><div class="message-content"><div class="message-meta"><span>{{ name(m.sender) }}</span><time>{{ time(m.time) }}</time></div><div class="message-bubble" :class="{ 'media-bubble': m.type !== 1 }"><template v-if="m.type === 1"><details v-if="m.reasoning" :key="`${m.id}-${m.state === 'succeeded'}`" class="ai-reasoning" :open="m.state === 'running'"><summary>思考过程</summary><p>{{ m.reasoning }}</p></details><p>{{ replyText(m) || (m.sender === 'ai' ? '…' : '') }}</p></template><ChatAttachment v-else :content="m.content" :type="m.type"/></div><div v-if="m.state && (m.sender === 'ai' || (self(m) && conversation.kind !== 'ai'))" class="send-status" :class="{ warning: ['failed','unknown'].includes(m.state) }"><Icon :name="m.state === 'succeeded' ? 'CheckCheck' : ['failed','unknown'].includes(m.state) ? 'AlertCircle' : 'LoaderCircle'" :size="12" :class="{ spin: ['sending','running','queued','processing'].includes(m.state) }"/><span :title="m.error">{{ stateText(m) }}</span><button v-if="self(m) && ['failed','unknown','processing'].includes(m.state)" @click="retry(m)">{{ m.state === 'failed' ? '重新发送' : conversation.kind === 'private' ? '同步确认' : '查询结果' }}</button></div><p v-if="m.error" class="message-error">{{ m.error }}</p></div></article>
    </template>
  </div>
  <button v-if="!atBottom" class="jump-bottom" @click="bottom"><Icon name="ChevronDown" :size="16"/>回到最新消息</button>
  <footer class="composer-area"><div class="composer" :class="{ disabled: !state.demo && state.connection !== 'online' }">
    <div class="composer-tools"><div class="flex items-center gap-1"><button class="icon-button" aria-label="表情" :aria-expanded="picker" @click="picker = !picker; mentionsOpen = false"><Icon name="Smile"/></button><button v-if="conversation.kind !== 'ai'" class="icon-button" aria-label="上传附件" :disabled="sending" @click="file?.click()"><Icon name="Paperclip"/></button><button v-if="conversation.kind === 'group'" class="icon-button" aria-label="提及群成员" @click="mentionPicker"><Icon name="AtSign"/></button></div><span class="composer-recipient">发送给 {{ conversation.name }}</span></div>
    <div v-if="picker" class="emoji-picker"><button v-for="emoji in ['😊','🤍','🌿','👍','🎉','☕','✨','🙌']" :key="emoji" :aria-label="`插入表情 ${emoji}`" @click="draft += emoji; picker = false; input?.focus()">{{ emoji }}</button></div>
    <div v-if="mentionsOpen" class="mention-picker"><button v-for="m in members.filter(m => m.id !== state.session?.uid)" :key="m.id" @click="mention(m)"><Avatar :name="m.name" size="xs"/>{{ m.name }}</button><p v-if="!members.length" class="muted small">正在获取群成员…</p></div>
    <div v-if="selectedMentions.length" class="mention-tags"><button v-for="id in selectedMentions" :key="id" @click="selectedMentions = selectedMentions.filter(x => x !== id)">@{{ members.find(m => m.id === id)?.name || name(id) }} <Icon name="X" :size="12"/></button></div>
    <textarea ref="input" v-model="draft" aria-label="消息内容" :placeholder="conversation.kind === 'ai' ? '写下你的问题或灵感…' : '写点什么，让对话继续…'" rows="2" @keydown="key" :disabled="sending"/>
    <div class="composer-bottom"><span>{{ sending ? '正在发送，请稍候…' : 'Enter 发送 · Shift + Enter 换行' }}</span><button class="send-button" aria-label="发送消息" :disabled="sending || !draft.trim() || (!state.demo && state.connection !== 'online')" @click="submit">发送<Icon :name="sending ? 'LoaderCircle' : 'ArrowUpRight'" :class="{ spin: sending }" :size="18"/></button></div>
  </div><p v-if="conversation.kind === 'ai'" class="composer-footnote">AI 的回答仅供参考。重要信息，请再确认一下。</p><p v-else class="composer-footnote">{{ conversation.kind === 'group' ? '共享此刻，也尊重彼此。' : '每一次真诚的交流，都值得被认真回应。' }}</p></footer>
  <input ref="file" class="sr-only" type="file" aria-label="选择附件" @change="attachment"/>
</section></template>

<style scoped>
.ai-reasoning { margin-bottom: 12px; padding-bottom: 10px; border-bottom: 1px solid var(--border, #e0e7e1); color: var(--text-secondary, #68746e); font-size: 12px; }
.ai-reasoning summary { cursor: pointer; padding: 3px 0; }
.ai-reasoning p { margin-top: 8px; white-space: pre-wrap; }
</style>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import Icon from './Icon.vue'
import Avatar from './Avatar.vue'
import ChatPane from './ChatPane.vue'
import ActionDialog from './ActionDialog.vue'
import DetailPanel from './DetailPanel.vue'
import AgentSettings from './AgentSettings.vue'
import { state, current, select, reload, reloadAI, logout, run, createAI, conversation } from '../lib/store'
import type { Conversation, Friend } from '../lib/protocol'
const dialog = ref(''), details = ref(false), menu = ref(false)
function keyboard(e: KeyboardEvent) {
  const target = e.target as HTMLElement
  if (e.key === '/' && !['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) && !target.isContentEditable && !dialog.value) { e.preventDefault(); document.querySelector<HTMLInputElement>('.sidebar-search input')?.focus() }
  if (e.key === 'Escape') { menu.value = false; details.value = false }
}
onMounted(() => { document.documentElement.dataset.compact = localStorage.getItem('qiyu.compact') || 'false'; window.addEventListener('keydown', keyboard) })
onUnmounted(() => window.removeEventListener('keydown', keyboard))
const filtered = computed(() => state.conversations.filter(c => (state.tab === 'ai' ? c.kind === 'ai' : c.kind !== 'ai') && (state.filter === 'all' || (state.filter === 'unread' ? !!c.unread : c.kind === 'group')) && `${c.name} ${c.preview}`.toLowerCase().includes(state.search.toLowerCase())).sort((a, b) => b.updated - a.updated))
const contacts = computed(() => state.friends.filter(f => `${f.name} ${f.remark}`.toLowerCase().includes(state.search.toLowerCase())))
const unread = computed(() => state.conversations.reduce((n, c) => n + (c.unread ? 1 : 0), 0))
function tab(t: typeof state.tab) { state.tab = t; state.search = ''; state.filter = 'all'; details.value = false; state.active = ''; if (t === 'ai') void run(reloadAI) }
function choose(c: Conversation) { details.value = false; void select(c) }
function openFriend(f: Friend) { let c = state.conversations.find(c => c.key === `private:${f.id}`); if (!c) { c = conversation(f.id, f.remark || f.name, 'private'); state.conversations.push(c) }; choose(c) }
function time(ts: number) { if (!ts) return ''; return new Date(ts).toLocaleDateString() === new Date().toLocaleDateString() ? new Date(ts).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }) : new Date(ts).toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' }) }
</script>
<template>
  <div class="workspace" :class="{ 'chat-open': !!current || state.active === 'ai-settings', 'detail-open': details }">
    <nav class="rail" aria-label="主导航">
      <a class="rail-brand" href="#" @click.prevent="tab('chats')" title="栖语"><img src="/mark.svg" alt="栖语"/></a>
      <div class="rail-links">
        <button class="rail-item" :class="{ active: state.tab === 'chats' }" @click="tab('chats')" aria-label="消息"><Icon name="MessageCircle" :size="23"/><span>消息</span><i v-if="unread" class="nav-dot"/></button>
        <button class="rail-item" :class="{ active: state.tab === 'contacts' }" @click="tab('contacts')" aria-label="通讯录"><Icon name="Users" :size="23"/><span>好友</span><i v-if="state.applications.length" class="nav-dot"/></button>
        <button class="rail-item" :class="{ active: state.tab === 'ai' }" @click="tab('ai')" aria-label="AI 对话"><Icon name="Sparkles" :size="23"/><span>AI</span></button>
      </div>
      <div class="rail-bottom"><button class="rail-item settings-nav" title="偏好设置" aria-label="偏好设置" @click="dialog = 'settings'"><Icon name="Settings2" :size="21"/></button><div class="rail-separator"/><button class="profile-nav" aria-label="个人资料" @click="dialog = 'profile'"><Avatar :name="state.profile.name || state.session?.account || '我'" :src="state.profile.photo" size="sm"/></button><button class="rail-item" aria-label="退出登录" title="退出登录" @click="dialog = 'logout'"><Icon name="LogOut" :size="19"/></button></div>
    </nav>
    <aside class="conversation-sidebar">
      <header class="sidebar-header"><div><span class="eyebrow">QIYU / 栖语</span><h1>{{ state.tab === 'contacts' ? '我的好友' : state.tab === 'ai' ? '灵感空间' : '消息' }}<span v-if="state.tab === 'chats' && unread" class="heading-count">{{ unread }}</span></h1></div><div class="plus-wrap"><button class="icon-button add-button" aria-label="新建" :aria-expanded="menu" @click="menu = !menu"><Icon name="Plus"/></button><div v-if="menu" class="dropdown"><button @click="dialog = 'friend'; menu = false"><Icon name="UserPlus" :size="17"/>添加好友</button><button @click="dialog = 'create-group'; menu = false"><Icon name="Users" :size="17"/>创建群聊</button><button @click="dialog = 'join-group'; menu = false"><Icon name="Hash" :size="17"/>申请入群</button></div></div></header>
      <div class="sidebar-search"><Icon name="Search" :size="17"/><input v-model="state.search" :placeholder="state.tab === 'contacts' ? '搜索好友' : '搜索会话或最近消息'" aria-label="搜索会话或好友"/><kbd>/</kbd></div>
      <div v-if="state.tab === 'chats'" class="filter-tabs"><button v-for="f in [{id:'all',name:'全部'}, {id:'unread',name:'未读'}, {id:'group',name:'群聊'}]" :key="f.id" :class="{ selected: state.filter === f.id }" @click="state.filter = f.id">{{ f.name }}</button></div>
      <div v-else-if="state.tab === 'contacts'" class="contacts-shortcut"><button @click="dialog = 'applications'"><span><Icon name="UserPlus" :size="18"/>新的朋友</span><b v-if="state.applications.length">{{ state.applications.length }}</b><Icon v-else name="ChevronRight" :size="16"/></button></div>
      <div v-else class="ai-actions"><button class="button primary" @click="run(createAI)"><Icon name="Plus" :size="17"/>开启新对话</button><button class="text-button" @click="dialog = 'restore-ai'">找回已有会话</button></div>
      <div class="list-caption"><span>{{ state.tab === 'contacts' ? '全部好友' : state.tab === 'ai' ? '我的对话' : '最近会话' }}</span><button class="icon-button" aria-label="刷新列表" :disabled="state.loading" @click="run(reload)"><Icon name="RefreshCw" :size="14" :class="{ spin: state.loading }"/></button></div>
      <div class="conversation-list" v-if="state.tab !== 'contacts'">
        <button v-if="state.tab === 'ai'" class="conversation-row agent-settings-pin" :class="{ selected: state.active === 'ai-settings' }" @click="state.active = 'ai-settings'; details = false"><Avatar name="智能体设置" kind="ai"/><span class="conversation-info"><strong>智能体设置</strong><span class="muted small block">头像、名称与角色设定</span></span><Icon name="Settings2" :size="18"/></button>
        <button v-for="c in filtered" :key="c.key" class="conversation-row" :class="{ selected: state.active === c.key }" @click="choose(c)"><Avatar :name="c.name" :src="c.avatar" :kind="c.kind" :online="c.kind === 'private' && c.online"/><span class="conversation-info"><span class="conversation-title"><strong>{{ c.name }}</strong><time>{{ time(c.updated) }}</time></span><span class="conversation-preview"><span>{{ c.preview || (c.kind === 'group' ? '和大家聊点什么吧' : '新消息') }}</span><b v-if="c.unread && c.kind === 'private'" class="badge">{{ c.unread > 99 ? '99+' : c.unread }}</b><i v-else-if="c.unread" class="unread-dot"/></span></span></button>
        <div v-if="!filtered.length" class="list-empty"><Icon :name="state.search ? 'Search' : 'MessageCircle'" :size="28"/><p>{{ state.search ? '没有找到相关会话' : state.filter === 'unread' ? '所有消息都看过了' : '还没有会话' }}</p><button class="text-button" @click="dialog = state.tab === 'ai' ? 'restore-ai' : 'friend'">{{ state.tab === 'ai' ? '找回会话' : '认识新的朋友' }}</button></div>
      </div>
      <div v-else class="conversation-list"><button v-for="f in contacts" :key="f.id" class="conversation-row" @click="openFriend(f)"><Avatar :name="f.remark || f.name" :src="f.avatar" :online="f.online"/><span class="conversation-info"><strong>{{ f.remark || f.name }}</strong><span class="muted small block truncate">{{ f.signature || '今天也值得好好聊聊' }}</span></span><Icon name="ChevronRight" :size="16" class="muted"/></button><div v-if="!contacts.length" class="list-empty"><Icon name="Users" :size="28"/><p>还没有找到朋友</p><button class="text-button" @click="dialog = 'friend'">添加好友</button></div></div>
      <footer class="sidebar-footer"><span class="connection-label"><i :class="state.connection"/>{{ state.demo ? '界面预览' : state.connection === 'online' ? '已连接，随时聊聊' : state.connection === 'connecting' ? '正在连接…' : '连接已断开，重连中' }}</span><span class="small muted">v1.0</span></footer>
    </aside>
    <main class="main-area">
      <div v-if="state.demo" class="demo-banner"><span>当前为界面预览 · 消息仅保存在本次页面</span><button @click="logout">登录真实账号 <Icon name="ArrowRight" :size="14"/></button></div>
      <div v-if="state.error" class="status-banner" role="status"><Icon name="WifiOff" :size="16"/><span>{{ state.error }}</span><button @click="run(reload)">重试</button></div>
      <AgentSettings v-if="state.active === 'ai-settings'" @back="state.active = ''"/>
      <ChatPane v-else-if="current" :key="current.key" :conversation="current" @details="details = !details"/>
      <div v-else class="welcome-pane"><div class="welcome-header"><span>一个舒服的角落，一段自在的对话。</span><span class="date-label">{{ new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric', weekday: 'long' }) }}</span></div><div class="welcome-content"><div class="welcome-symbol"><Icon :name="state.tab === 'ai' ? 'Sparkles' : 'MessageCircle'" :size="53"/><span class="symbol-spark"><Icon name="Sparkles" :size="23"/></span></div><span class="eyebrow">MAKE ROOM FOR CONNECTION</span><h2>{{ state.tab === 'ai' ? '给想法，一个回响。' : '此刻，想和谁聊聊？' }}</h2><p>{{ state.tab === 'ai' ? '整理思路、探索灵感，或为一天写下小结。' : '挑选左侧的一个会话，接着分享想法。' }}<br/>{{ state.tab === 'ai' ? '开启一段属于你的 AI 对话。' : '也可以从一句「你好」开始，认识新朋友。' }}</p><button class="button primary" @click="state.tab === 'ai' ? run(createAI) : dialog = 'friend'"><Icon :name="state.tab === 'ai' ? 'Plus' : 'UserPlus'" :size="18"/>{{ state.tab === 'ai' ? '开启新对话' : '添加好友' }}<Icon name="ArrowUpRight" :size="17"/></button><div class="welcome-quote"><span></span></div></div><footer class="welcome-footer"><Icon name="Leaf" :size="16"/>保持好奇，保持联系。<span>QIYU — STAY CONNECTED</span></footer></div>
    </main>
    <DetailPanel v-if="details && current" :key="current.key" :conversation="current" @close="details = false"/>
    <ActionDialog v-if="dialog" :kind="dialog" @close="dialog = ''"/>
  </div>
</template>

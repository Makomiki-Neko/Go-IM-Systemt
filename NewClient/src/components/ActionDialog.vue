<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import Modal from './Modal.vue'
import Icon from './Icon.vue'
import Avatar from './Avatar.vue'
import { state, run, logout, groupAction, friendAction, reload, updateProfile, restoreAI, toast, select, clearCache } from '../lib/store'
import { array, friendFrom, str, type Friend } from '../lib/protocol'
const props = defineProps<{ kind: string }>()
const emit = defineEmits<{ close: [] }>()
const titles: Record<string, string> = { friend: '认识新的朋友', 'create-group': '给大家，一个相聚的地方', 'join-group': '加入新的群聊', applications: '新的朋友', profile: '关于我', settings: '偏好设置', logout: '暂时离开一下？', 'restore-ai': '找回 AI 会话' }
const title = computed(() => titles[props.kind] || '')
const busy = ref(false), query = ref(''), results = ref<Friend[]>([]), searched = ref(false), message = ref('你好，很高兴认识你。'), groupName = ref(''), notice = ref(''), id = ref('')
const profile = reactive({ ...state.profile }), picture = ref<File>(), picturePreview = ref('')
const compact = ref(localStorage.getItem('qiyu.compact') === 'true')
async function act(fn: () => Promise<unknown>, success: string, close = true) {
  if (busy.value) return
  busy.value = true; const ok = await run(fn, success); busy.value = false
  if (ok && close) emit('close')
}
async function search() {
  if (!query.value.trim()) return
  await act(async () => { if (state.demo) results.value = state.friends.filter(f => f.name.includes(query.value)); else { const d = await friendAction(`search?search_name=${encodeURIComponent(query.value)}&page=1&size=50`, {}); results.value = array(d.list).map(friendFrom) }; searched.value = true }, '', false)
}
async function create() {
  await act(async () => { const d = await groupAction('create', { name: groupName.value.trim(), notice: notice.value, avatar: '' }); await reload(); const c = state.conversations.find(c => c.id === str(d.group_id) && c.kind === 'group'); if (c) await select(c) }, '群聊已创建，可在详情中邀请好友')
}
function avatarFile(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  if (!f) return
  if (!f.type.startsWith('image/') || f.size > 5 * 1024 * 1024) { toast('请选择 5 MB 以内的图片', true); return }
  picture.value = f
  const reader = new FileReader(); reader.onload = () => { picturePreview.value = String(reader.result) }; reader.readAsDataURL(f)
}
function appearance() { document.documentElement.dataset.compact = String(compact.value); localStorage.setItem('qiyu.compact', String(compact.value)) }
</script>
<template><Modal :title="title" :subtitle="kind === 'friend' ? '搜索账号或昵称，和有趣的人相遇。' : kind === 'profile' ? '让朋友们更容易认出你。' : undefined" @close="emit('close')">
  <div class="modal-body">
    <template v-if="kind === 'friend'"><form class="search-form" @submit.prevent="search"><input v-model="query" placeholder="输入账号或昵称" aria-label="搜索账号或昵称" required/><button class="button primary" :disabled="busy">搜索</button></form><label class="form-label">打招呼的内容<input v-model="message" maxlength="50"/></label><div class="search-results"><div v-for="f in results" :key="f.id" class="person-row"><Avatar :name="f.name" :src="f.avatar"/><div><strong>{{ f.name }}</strong><p>{{ f.signature || `ID ${f.id}` }}</p></div><button class="button secondary small-button" :disabled="busy || f.id === state.session?.uid || state.friends.some(x => x.id === f.id)" @click="act(() => friendAction('request', { friend_id: f.id, msg: message }), '好友申请已发送', false)">{{ f.id === state.session?.uid ? '我自己' : state.friends.some(x => x.id === f.id) ? '已是好友' : '加好友' }}</button></div><p v-if="searched && !results.length" class="empty-note">没有找到相关用户，换个关键词试试。</p></div></template>
    <form v-else-if="kind === 'create-group'" class="form-stack" @submit.prevent="create"><label>群聊名称<input v-model="groupName" required maxlength="50" placeholder="例如：周末散步小组"/></label><label>群公告 <span class="muted">选填</span><textarea v-model="notice" maxlength="255" rows="3" placeholder="写下一点共同的期待…"/></label><p class="field-help">创建后可以邀请好友，或分享群 ID 让朋友申请加入。</p><button class="button primary" :disabled="busy">{{ busy ? '创建中…' : '创建群聊' }}<Icon name="ArrowRight" :size="17"/></button></form>
    <form v-else-if="kind === 'join-group'" class="form-stack" @submit.prevent="act(() => groupAction('apply', { group_id: id, message }), '入群申请已发送')"><label>群 ID<input v-model="id" required pattern="[0-9]+" inputmode="numeric" placeholder="向朋友询问群 ID"/></label><label>申请留言<textarea v-model="message" maxlength="255" rows="3"/></label><button class="button primary" :disabled="busy">发送入群申请</button></form>
    <template v-else-if="kind === 'applications'"><p class="field-help">收到的好友申请会显示在这里。群申请请到群聊详情中处理。</p><div v-for="f in state.applications" :key="f.id" class="person-row"><Avatar :name="f.name" :src="f.avatar"/><div><strong>{{ f.name }}</strong><p>{{ f.remark || '希望成为你的好友' }}</p></div><button class="button primary small-button" :disabled="busy" @click="act(async () => { await friendAction('request_handle', { request_id: '0', sender_id: f.id, user_id: state.session?.uid, accept: true }); await reload() }, '已添加为好友', false)">接受</button></div><div v-if="!state.applications.length" class="list-empty"><Icon name="UserPlus" :size="32"/><p>暂时没有新的申请</p></div><p v-if="state.applications.length" class="field-help">暂不支持在此拒绝申请：当前服务未返回拒绝操作所需的申请编号。</p></template>
    <form v-else-if="kind === 'profile'" class="form-stack" @submit.prevent="act(() => updateProfile(profile, picture), '资料已保存')"><div class="profile-preview"><img v-if="picturePreview" :src="picturePreview" class="profile-picture" alt="新头像预览"/><Avatar v-else :name="profile.name || state.session?.account || '我'" :src="profile.photo" size="xl"/><label class="button secondary small-button"><Icon name="Camera" :size="16"/>更换头像<input type="file" accept="image/*" class="sr-only" @change="avatarFile"/></label></div><label>昵称<input v-model="profile.name" required maxlength="100"/></label><label>个人签名<textarea v-model="profile.signature" rows="2" maxlength="250" placeholder="一句话，介绍此刻的你。"/></label><div class="two-columns"><label>性别<select v-model="profile.gender"><option value="">不公开</option><option>男</option><option>女</option><option>其他</option></select></label><label>生日<input type="date" :value="profile.birthday ? new Date(profile.birthday * 1000).toISOString().slice(0,10) : ''" @input="profile.birthday = Date.parse(($event.target as HTMLInputElement).value) / 1000 || 0"/></label></div><p class="field-help">账号：{{ state.session?.account }} · ID：{{ state.session?.uid }}</p><button class="button primary" :disabled="busy">保存资料</button><button type="button" class="text-button" @click="logout(); emit('close')">{{ state.demo ? '退出预览' : '退出登录' }}</button></form>
    <template v-else-if="kind === 'settings'"><div class="setting-row"><div><strong>紧凑会话列表</strong><p>减少列表留白，在一屏看到更多会话。</p></div><input v-model="compact" type="checkbox" class="toggle" aria-label="紧凑会话列表" @change="appearance"/></div><div class="setting-row"><div><strong>消息与隐私</strong><p>聊天缓存保存在当前浏览器，并按账号隔离。公共设备使用后，建议清除本地记录。</p></div></div><button class="button secondary" @click="clearCache(); toast('已清除持久缓存；当前页面记录将在退出后释放'); emit('close')">清除已保存的聊天缓存</button><div class="settings-about"><img src="/mark.svg" alt=""/><div><strong>栖语 QIYU</strong><p>v1.0 · 让交流，自然发生。</p></div></div></template>
    <template v-else-if="kind === 'logout'"><p class="muted">{{ state.demo ? '退出预览后，即可使用自己的账号登录。' : '退出后将断开当前连接，本设备的聊天记录仍会保留。' }}</p><div class="modal-actions"><button class="button secondary" @click="emit('close')">再聊一会儿</button><button class="button primary" @click="logout(); emit('close')">{{ state.demo ? '退出预览' : '退出登录' }}</button></div></template>
    <form v-else-if="kind === 'restore-ai'" class="form-stack" @submit.prevent="act(() => restoreAI(id), '已打开会话')"><p class="field-help">输入属于你自己的会话 ID。新建会话会自动保存在此浏览器；其他设备上的会话可通过 ID 找回。</p><label>会话 ID<input v-model="id" required pattern="[0-9]+" inputmode="numeric"/></label><button class="button primary" :disabled="busy">打开会话</button></form>
  </div>
</Modal></template>

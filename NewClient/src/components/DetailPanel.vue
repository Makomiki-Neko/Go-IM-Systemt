<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from './Icon.vue'
import Avatar from './Avatar.vue'
import Modal from './Modal.vue'
import { state, groupAction, groupList, friendAction, reload, run, toast, deleteAI } from '../lib/store'
import { array, object, str, num, type Conversation, type JsonObject } from '../lib/protocol'
const props = defineProps<{ conversation: Conversation }>()
const emit = defineEmits<{ close: [] }>()
const members = ref<JsonObject[]>([]), requests = ref<JsonObject[]>([]), info = ref<JsonObject>({}), loading = ref(false), editing = ref(false), inviting = ref(false), confirm = ref(''), name = ref(props.conversation.name), notice = ref(''), nickname = ref(''), remark = ref(props.conversation.name), selected = ref<string[]>([]), memberPage = ref(1), moreMembers = ref(false), error = ref('')
const myRole = computed(() => members.value.find(m => str(m.user_id) === state.session?.uid)?.role)
const owner = computed(() => myRole.value === 'MEMBER_ROLE_OWNER')
const manager = computed(() => owner.value || myRole.value === 'MEMBER_ROLE_ADMIN')
const label = (id: string) => id === state.session?.uid ? state.profile.name || '我' : state.friends.find(f => f.id === id)?.name || `成员 ${id.slice(-5)}`
const roleLabel = (role: unknown) => role === 'MEMBER_ROLE_OWNER' ? '群主' : role === 'MEMBER_ROLE_ADMIN' ? '管理员' : ''
const invitees = computed(() => state.friends.filter(f => !members.value.some(m => str(m.user_id) === f.id)))
async function load() {
  if (props.conversation.kind !== 'group') return
  if (state.demo) {
    info.value = { name: props.conversation.name, member_count: 4, notice: '在这里分享有趣的事，也欢迎那些还没成形的想法。慢慢聊，不着急。' }
    members.value = [{ user_id: state.session!.uid, role: 'MEMBER_ROLE_OWNER', group_nick: '' }, ...state.friends.map(f => ({ user_id: f.id, role: 'MEMBER_ROLE_MEMBER', group_nick: '' }))]; notice.value = str(info.value.notice); return
  }
  loading.value = true; error.value = ''
  const ok = await run(async () => {
    const [a, b] = await Promise.all([groupAction('info', { group_id: props.conversation.id }), groupList('members', props.conversation.id)])
    info.value = object(a.info); name.value = str(info.value.name); notice.value = str(info.value.notice)
    members.value = b; memberPage.value = 1; moreMembers.value = false
    if (manager.value) requests.value = await groupList('requests', props.conversation.id)
  })
  if (!ok) error.value = '详情加载失败，请重试'
  loading.value = false
}
async function action(action: string, body: JsonObject = {}, success = '已更新') {
  loading.value = true
  const ok = await run(async () => { await groupAction(action, { group_id: props.conversation.id, ...body }); await reload(); if (!['quit','dismiss'].includes(action)) await load() }, success)
  loading.value = false
  if (ok) { confirm.value = ''; editing.value = false; inviting.value = false; if (['quit','dismiss'].includes(action)) emit('close') }
}
async function copyID() { await run(async () => { await navigator.clipboard.writeText(props.conversation.id) }, 'ID 已复制') }
async function more() { await run(async () => { const d = await groupAction('members', { group_id: props.conversation.id, page: { page: memberPage.value + 1, size: 200 } }); members.value.push(...array(d.list).map(object)); memberPage.value++; moreMembers.value = members.value.length < num(object(d.page).total) }) }
onMounted(load)
</script>
<template><aside class="detail-panel"><header><h3>{{ conversation.kind === 'group' ? '群聊空间' : conversation.kind === 'ai' ? '对话详情' : '好友资料' }}</h3><button class="icon-button" aria-label="关闭会话详情" @click="emit('close')"><Icon name="PanelRightClose" :size="19"/></button></header>
  <div class="detail-scroll"><div class="detail-identity"><Avatar :name="conversation.name" :src="conversation.avatar" :kind="conversation.kind" size="xl"/><h2>{{ conversation.name }}</h2><button class="id-copy" @click="copyID">ID {{ conversation.id }}<Icon name="Copy" :size="12"/></button><span v-if="conversation.kind === 'group'" class="detail-member-count">{{ info.member_count || members.length }} 位成员，一起分享</span></div>
    <p v-if="error" class="inline-error">{{ error }}<button @click="load">重试</button></p>
    <template v-if="conversation.kind === 'group'"><section class="detail-section"><div class="section-heading"><h4>群公告</h4><button v-if="manager" class="icon-button" aria-label="编辑群资料" @click="editing = true"><Icon name="Pencil" :size="15"/></button></div><p class="group-notice">{{ info.notice || '暂时没有公告，留一点空间给新的故事。' }}</p></section><section class="detail-section"><div class="section-heading"><h4>群成员 <span>{{ info.member_count || members.length }}</span></h4><button v-if="manager" class="icon-button" aria-label="邀请好友" @click="inviting = true"><Icon name="UserPlus" :size="17"/></button></div><div class="member-grid"><div v-for="m in members" :key="str(m.user_id)" class="member-tile"><Avatar :name="str(m.group_nick) || label(str(m.user_id))" size="sm"/><span>{{ str(m.group_nick) || label(str(m.user_id)) }}</span><small v-if="roleLabel(m.role)">{{ roleLabel(m.role) }}</small><div v-if="manager && str(m.user_id) !== state.session?.uid && m.role !== 'MEMBER_ROLE_OWNER'" class="member-actions"><button v-if="owner" :disabled="loading" @click="action('role', { target_id: str(m.user_id), new_role: m.role === 'MEMBER_ROLE_ADMIN' ? 'MEMBER_ROLE_MEMBER' : 'MEMBER_ROLE_ADMIN' })">{{ m.role === 'MEMBER_ROLE_ADMIN' ? '撤管理员' : '设管理员' }}</button><button v-if="owner || m.role === 'MEMBER_ROLE_MEMBER'" :disabled="loading" @click="confirm = `remove:${str(m.user_id)}`">移出</button></div></div></div><button v-if="moreMembers" class="text-button" @click="more">加载更多成员</button></section>
    <section v-if="manager && requests.length" class="detail-section"><h4>待处理申请 <span>{{ requests.length }}</span></h4><div v-for="r in requests" :key="str(r.id)" class="join-request"><span>{{ label(str(r.user_id)) }}</span><button class="icon-button" aria-label="同意入群" :disabled="loading" @click="action('process', { applicant_id: str(r.user_id), accept: true }, '已同意入群')"><Icon name="Check" :size="17"/></button><button class="icon-button" aria-label="拒绝入群" :disabled="loading" @click="action('process', { applicant_id: str(r.user_id), accept: false }, '已拒绝申请')"><Icon name="X" :size="17"/></button></div></section>
    <section class="detail-section"><h4>我在本群的昵称</h4><form class="inline-form" @submit.prevent="action('nick', { group_nick: nickname })"><input v-model="nickname" placeholder="设置一个群昵称" maxlength="50" required aria-label="群昵称"/><button class="icon-button" :disabled="loading" aria-label="保存群昵称"><Icon name="Check" :size="18"/></button></form></section><div class="detail-note"><Icon name="Leaf" :size="16"/><p>只能查看本次入群后的消息。重新加入，将从新的故事开始。</p></div><button class="danger-link" @click="confirm = owner ? 'dismiss' : 'quit'">{{ owner ? '解散群聊' : '退出群聊' }}</button></template>
    <template v-else-if="conversation.kind === 'private'"><section class="detail-section"><h4>个人签名</h4><p class="group-notice">{{ conversation.subtitle || '这个朋友还没有留下签名。' }}</p></section><section class="detail-section"><h4>备注名称</h4><form class="inline-form" @submit.prevent="run(async () => { await friendAction('remark', { friend_id: conversation.id, remark }); await reload() }, '备注已保存')"><input v-model="remark" maxlength="50" aria-label="好友备注"/><button class="icon-button" aria-label="保存备注"><Icon name="Check" :size="18"/></button></form></section><button class="danger-link" @click="confirm = 'delete-friend'">删除好友</button></template>
    <template v-else><section class="detail-section"><h4>属于你的灵感空间</h4><p class="group-notice">AI 会参考当前会话的上下文，较长的对话会自动整理为摘要。</p></section><div class="detail-note"><Icon name="CircleHelp" :size="18"/><p>当前不支持跨会话长期记忆。会话列表会从服务器同步。</p></div><button class="danger-link" @click="confirm = 'delete-ai'">删除 AI 对话</button></template>
  </div>
  <Modal v-if="editing" title="编辑群资料" @close="editing = false"><form class="modal-body form-stack" @submit.prevent="action('update', { name, notice })"><label>群名<input v-model="name" maxlength="50" required/></label><label>群公告<textarea v-model="notice" rows="4" maxlength="255"/></label><p class="field-help">留空的字段将保持原值。</p><button class="button primary" :disabled="loading">保存修改</button></form></Modal>
  <Modal v-if="inviting" title="邀请好友，一起聊聊" @close="inviting = false"><form class="modal-body form-stack" @submit.prevent="action('invite', { invitee_ids: selected }, '已邀请加入群聊')"><label v-for="f in invitees" :key="f.id" class="invite-row"><input v-model="selected" type="checkbox" :value="f.id"/><Avatar :name="f.name" size="sm"/><span>{{ f.remark || f.name }}</span></label><p v-if="!invitees.length" class="empty-note">你的好友都已在当前成员列表中。</p><button class="button primary" :disabled="loading || !selected.length">邀请 {{ selected.length }} 位好友</button></form></Modal>
  <Modal v-if="confirm" :title="confirm === 'delete-ai' ? '确认删除 AI 对话？' : confirm.startsWith('remove') ? '移出这位成员？' : confirm === 'dismiss' ? '确认解散群聊？' : confirm === 'quit' ? '确认退出群聊？' : '确认删除好友？'" @close="confirm = ''"><div class="modal-body"><p class="muted">{{ confirm === 'delete-ai' ? '删除后将无法继续发送或拉取本会话，服务器保留软删除记录。' : confirm === 'dismiss' ? '所有成员将无法继续发送或从服务器拉取本群消息。' : confirm === 'quit' ? '退出后将无法从服务器拉取群消息，本地缓存保留。' : confirm.startsWith('remove') ? '该成员将失去本群的消息访问权限。' : '删除后将无法继续向这位好友发送私聊消息。' }}</p><div class="modal-actions"><button class="button secondary" @click="confirm = ''">取消</button><button class="button danger" :disabled="loading" @click="confirm === 'delete-ai' ? run(async () => { await deleteAI(conversation); emit('close') }, '已删除对话') : confirm === 'delete-friend' ? run(async () => { await friendAction('delete', { friend_id: conversation.id }); await reload(); emit('close') }, '已删除好友') : action(confirm.startsWith('remove') ? 'remove' : confirm, confirm.startsWith('remove') ? { target_id: confirm.split(':')[1] } : {})">确认</button></div></div></Modal>
</aside></template>

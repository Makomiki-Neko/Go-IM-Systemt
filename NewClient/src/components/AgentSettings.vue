<script setup lang="ts">
import { onMounted, ref } from 'vue'
import Avatar from './Avatar.vue'
import Icon from './Icon.vue'
import { agent, loadAgent, saveAgent, run, state } from '../lib/store'
import { request } from '../lib/transport'
import { str } from '../lib/protocol'
defineEmits<{ back: [] }>()
const name = ref(''), description = ref(''), avatar = ref(''), busy = ref(false), ready = ref(false), file = ref<HTMLInputElement>()
async function load() {
  busy.value = true
  ready.value = await run(async () => { await loadAgent(); name.value = agent.name; description.value = agent.description; avatar.value = agent.avatar })
  busy.value = false
}
async function save() {
  busy.value = true
  await run(() => saveAgent(name.value, description.value, avatar.value), '智能体已保存，将用于新对话')
  busy.value = false
}
async function upload(e: Event) {
  const input = e.target as HTMLInputElement, chosen = input.files?.[0]; input.value = ''
  if (!chosen) return
  busy.value = true
  await run(async () => {
    if (state.demo) throw new Error('请登录后上传头像')
    if (chosen.size > 5 * 1024 * 1024) throw new Error('头像不能超过 5 MB')
    const data = new FormData(); data.append('photo', chosen)
    avatar.value = str((await request('/ai/agent/avatar', data)).path)
  })
  busy.value = false
}
onMounted(load)
</script>
<template><section class="agent-settings">
  <header class="chat-header"><button class="icon-button mobile-back" aria-label="返回 AI 列表" @click="$emit('back')"><Icon name="ArrowLeft"/></button><div class="chat-title"><h2>智能体设置</h2><p>为每一次新对话，设定一个熟悉的角色。</p></div></header>
  <div class="agent-settings-body">
    <div class="agent-preview"><Avatar :name="name || '我的智能体'" :src="avatar" kind="ai" size="xl"/><div><span class="eyebrow">YOUR AI COMPANION</span><h3>{{ name || '我的智能体' }}</h3><p class="muted">一个角色，许多段故事。</p></div></div>
    <form v-if="ready" class="form-stack" @submit.prevent="save">
      <div class="flex items-center gap-3"><button type="button" class="button secondary" :disabled="busy" @click="file?.click()"><Icon name="Image" :size="16"/>选择头像</button><button v-if="avatar" type="button" class="text-button" :disabled="busy" @click="avatar = ''">移除头像</button><span class="field-help">JPG、PNG、GIF，最多 5 MB</span></div>
      <label>角色名称<input v-model="name" maxlength="80" required :disabled="busy" placeholder="例如：灯塔书屋的守夜人"/></label>
      <label>角色描述<textarea v-model="description" rows="10" required :disabled="busy" placeholder="描述角色的身份、性格、说话方式，以及你们故事的背景…"/></label>
      <p class="field-help">角色描述会作为新会话的初始设定。保存修改不会改变已有会话的角色；每个账号暂支持一个自定义智能体。描述最多 16000 字节。</p>
      <div><button class="button primary" :disabled="busy || !name.trim() || !description.trim()">{{ busy ? '正在处理…' : '保存智能体' }}<Icon name="Check" :size="17"/></button></div>
    </form>
    <button v-else class="button secondary" :disabled="busy" @click="load">{{ busy ? '正在加载…' : '重新加载设置' }}</button>
  </div>
  <input ref="file" class="sr-only" type="file" accept="image/jpeg,image/png,image/gif" aria-label="选择智能体头像" @change="upload"/>
</section></template>

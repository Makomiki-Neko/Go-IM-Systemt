<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { request } from '../lib/transport'
import { chatFilePath } from '../lib/media'
import { str, num } from '../lib/protocol'
const props = defineProps<{ content: string; type: number }>()
const root = ref<HTMLElement>(), url = ref(''), loading = ref(false), error = ref('')
let expiresAt = 0, generation = 0, recovery = 0, visible = false
let pending: Promise<string> | undefined, observer: IntersectionObserver | undefined
async function authorize(force = false): Promise<string> {
  if (!force && url.value && Date.now() < expiresAt - 15000) return url.value
  if (pending) return pending
  const version = generation
  loading.value = true; error.value = ''
  const work = (async () => {
    const path = chatFilePath(props.content)
    if (!path) throw new Error('无法识别附件路径')
    const result = await request('/file/download-url', { path })
    const signed = new URL(str(result.url))
    if (!['http:', 'https:'].includes(signed.protocol) || num(result.expires_at) * 1000 <= Date.now()) throw new Error('服务器返回了无效的下载链接')
    if (version !== generation) return ''
    expiresAt = num(result.expires_at) * 1000; url.value = signed.href
    return signed.href
  })()
  pending = work
  try { return await work }
  catch (e) { if (version === generation) { error.value = e instanceof Error ? e.message : '附件加载失败'; url.value = '' }; return '' }
  finally { if (version === generation) { loading.value = false; pending = undefined } }
}
async function mediaError() {
  if (recovery++ === 0) { await authorize(true); return }
  error.value = '附件加载失败，文件可能已删除或存储服务不可达'
}
async function retry() { recovery = 0; await authorize(true) }
async function open() {
  // Open synchronously so browsers permit a new tab, then resolve/refresh its URL.
  const tab = window.open('about:blank', '_blank')
  if (tab) tab.opener = null
  const signed = await authorize()
  if (!signed) { tab?.close(); return }
  if (tab) tab.location.replace(signed)
  else { const a = document.createElement('a'); a.href = signed; a.target = '_blank'; a.rel = 'noopener noreferrer'; a.click() }
}
watch(() => [props.content, props.type], () => {
  generation++; pending = undefined; url.value = ''; error.value = ''; loading.value = false; expiresAt = 0; recovery = 0
  if (visible && props.type >= 2 && props.type <= 4) void authorize()
})
onMounted(() => {
  observer = new IntersectionObserver(entries => {
    if (entries.some(e => e.isIntersecting)) { visible = true; observer?.disconnect(); if (props.type >= 2 && props.type <= 4) void authorize() }
  }, { rootMargin: '200px' })
  if (root.value) observer.observe(root.value)
})
onBeforeUnmount(() => { generation++; observer?.disconnect() })
</script>
<template><span ref="root">
  <span v-if="error" role="alert">{{ error }} <button class="text-button" @click="retry">重试</button></span>
  <span v-else-if="loading" role="status">正在获取附件链接…</span>
  <template v-else-if="url">
    <a v-if="type === 2" :href="url" target="_blank" rel="noopener noreferrer" @click.prevent="open"><img class="message-image" :src="url" alt="聊天图片" loading="lazy" @error="mediaError"/></a>
    <audio v-else-if="type === 3" controls preload="none" :src="url" @error="mediaError"/>
    <video v-else-if="type === 4" controls preload="metadata" :src="url" @error="mediaError"/>
    <a v-else class="file-message" :href="url" target="_blank" rel="noopener noreferrer" @click.prevent="open">下载 {{ content.split('/').pop() }}</a>
  </template>
  <button v-else class="text-button" @click="type > 4 ? open() : retry()">{{ type > 4 ? '下载附件' : '加载附件' }} · {{ content.split('/').pop() }}</button>
</span></template>

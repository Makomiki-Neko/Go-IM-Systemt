<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import Icon from './Icon.vue'
defineProps<{ title: string; subtitle?: string; wide?: boolean }>()
const emit = defineEmits<{ close: [] }>()
const panel = ref<HTMLElement>()
let previous: HTMLElement | null = null
function key(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
  if (e.key !== 'Tab') return
  const nodes = Array.from(panel.value?.querySelectorAll<HTMLElement>('button:not(:disabled),input:not(:disabled),select,textarea,a[href],[tabindex="0"]') || []).filter(el => el.offsetParent !== null)
  const first = nodes[0], last = nodes.at(-1)
  if (e.shiftKey && document.activeElement === first) { e.preventDefault(); last?.focus() }
  if (!e.shiftKey && document.activeElement === last) { e.preventDefault(); first?.focus() }
}
onMounted(() => { previous = document.activeElement as HTMLElement; panel.value?.focus(); document.addEventListener('keydown', key) })
onUnmounted(() => { document.removeEventListener('keydown', key); previous?.focus() })
</script>
<template><Teleport to="body"><div class="modal-backdrop" @click.self="emit('close')"><section ref="panel" class="modal" :class="{ wide }" role="dialog" aria-modal="true" :aria-label="title" tabindex="-1">
  <header class="modal-head"><div><h2>{{ title }}</h2><p v-if="subtitle">{{ subtitle }}</p></div><button class="icon-button" aria-label="关闭弹窗" @click="emit('close')"><Icon name="X" /></button></header>
  <slot />
</section></div></Teleport></template>

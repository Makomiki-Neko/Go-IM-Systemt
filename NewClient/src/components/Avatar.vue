<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import Icon from './Icon.vue'
import { safeURL } from '../lib/protocol'
const props = withDefaults(defineProps<{ name: string; src?: string; kind?: string; size?: string; online?: boolean }>(), { size: 'md', kind: 'private' })
const failed = ref(false)
watch(() => props.src, () => { failed.value = false })
const palette = computed(() => ['sage', 'sand', 'clay', 'blue'][(props.name.codePointAt(0) || 0) % 4])
const image = computed(() => props.src && !props.src.endsWith('default.jpg') ? safeURL(props.src, import.meta.env.VITE_FILER_BASE || '/filer') : '')
</script>
<template><span class="avatar" :class="[size, kind === 'ai' ? 'ai-avatar' : palette]">
  <img v-if="image && !failed" :src="image" :alt="name" @error="failed = true" />
  <Icon v-else-if="kind === 'ai'" name="Sparkles" :size="size === 'xl' ? 32 : 21" />
  <Icon v-else-if="kind === 'group'" name="Users" :size="size === 'xl' ? 32 : 22" />
  <span v-else>{{ name.slice(0, 1) || '友' }}</span>
  <i v-if="online" class="online-dot" />
</span></template>

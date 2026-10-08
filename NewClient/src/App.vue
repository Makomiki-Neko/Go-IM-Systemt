<script setup lang="ts">
import { onMounted } from 'vue'
import AuthView from './components/AuthView.vue'
import Workspace from './components/Workspace.vue'
import Icon from './components/Icon.vue'
import { state, start } from './lib/store'
onMounted(() => { if (state.session) void start() })
</script>
<template><Workspace v-if="state.session"/><AuthView v-else/>
  <div class="toast-stack" aria-live="polite"><div v-for="item in state.toasts" :key="item.id" class="toast" :class="{ error: item.error }"><Icon :name="item.error ? 'AlertCircle' : 'CheckCircle2'" :size="19"/><span>{{ item.text }}</span><button class="icon-button" aria-label="关闭提示" @click="state.toasts = state.toasts.filter(t => t.id !== item.id)"><Icon name="X" :size="16"/></button></div></div>
</template>

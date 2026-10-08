<script setup lang="ts">
import { ref } from 'vue'
import Icon from './Icon.vue'
import { login, register, enterDemo, explain } from '../lib/store'
const mode = ref<'login' | 'register'>('login')
const account = ref(''), password = ref(''), email = ref(''), confirmation = ref('')
const visible = ref(false), busy = ref(false), error = ref(''), success = ref('')
async function submit() {
  error.value = ''; success.value = ''
  if (mode.value === 'register' && password.value !== confirmation.value) { error.value = '两次输入的密码不一致'; return }
  busy.value = true
  try {
    if (mode.value === 'register') { await register(account.value.trim(), password.value, email.value.trim()); mode.value = 'login'; success.value = '账号已创建，登录后开始新的交流。' }
    else await login(account.value.trim(), password.value)
  } catch (e) { error.value = explain(e) }
  finally { busy.value = false }
}
</script>
<template>
  <main class="auth-page">
    <section class="auth-story">
      <a class="wordmark" href="#" aria-label="栖语首页"><img src="/mark.svg" alt=""/><span>栖语 <small>QIYU</small></span></a>
      <div class="story-copy"><span class="eyebrow">A LITTLE CLOSER, EVERY DAY</span><h1>让交流，<br/>自然发生<span class="accent-stop">。</span></h1><p>分享灵感，也分享平凡的日常。</p></div>
      <div class="story-art" aria-hidden="true"><div class="orbit orbit-one"></div><div class="orbit orbit-two"></div><div class="art-leaf"><Icon name="Leaf" :size="78" /></div><div class="art-note note-one"><span class="little-avatar">予</span><div>今天有什么新鲜事？<small>分享，让灵感生长</small></div></div><div class="art-note note-two"><Icon name="MessageCircle" :size="26"/><span>每一句，都有回响。</span><i></i></div><span class="art-caption">STAY CURIOUS. STAY CONNECTED.</span></div>
      <footer><span>……</span><span>© {{ new Date().getFullYear() }} QIYU</span></footer>
    </section>
    <section class="auth-form-area">
      <div class="auth-top"><span>{{ mode === 'login' ? '还没有账号？' : '已经有账号？' }}</span><button class="text-button" @click="mode = mode === 'login' ? 'register' : 'login'; error = ''; success = ''">{{ mode === 'login' ? '加入栖语' : '去登录' }} <Icon name="ArrowUpRight" :size="16"/></button></div>
      <div class="auth-form-wrap"><span class="eyebrow">YOUR SPACE TO CONNECT</span><h2>{{ mode === 'login' ? '很高兴，又见面了' : '从一句你好开始' }}</h2><p class="auth-intro">{{ mode === 'login' ? '登录栖语，接着聊聊那些有趣的事。' : '创建账号，让彼此的世界靠近一点。' }}</p>
        <form @submit.prevent="submit" class="form-stack">
          <label>账号<div class="input-icon"><Icon name="UserRound" :size="18"/><input v-model="account" autocomplete="username" placeholder="请输入你的账号" required maxlength="100" :disabled="busy"/></div></label>
          <label v-if="mode === 'register'">邮箱<div class="input-icon"><Icon name="Mail" :size="18"/><input v-model="email" type="email" autocomplete="email" placeholder="you@example.com" required maxlength="100" :disabled="busy"/></div></label>
          <label>密码<div class="input-icon"><Icon name="LockKeyhole" :size="18"/><input v-model="password" :type="visible ? 'text' : 'password'" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" placeholder="请输入密码" required :minlength="mode === 'register' ? 6 : 1" maxlength="72" :disabled="busy"/><button type="button" class="icon-button" :aria-label="visible ? '隐藏密码' : '显示密码'" @click="visible = !visible"><Icon :name="visible ? 'EyeOff' : 'Eye'" :size="18"/></button></div></label>
          <label v-if="mode === 'register'">确认密码<input v-model="confirmation" type="password" autocomplete="new-password" placeholder="再次输入密码" required :disabled="busy"/></label>
          <p v-if="error" class="inline-error" role="alert"><Icon name="AlertCircle" :size="17"/>{{ error }}</p>
          <p v-if="success" class="inline-success" role="status">{{ success }}</p>
          <button class="button primary auth-submit" :disabled="busy"><Icon v-if="busy" name="LoaderCircle" class="spin" :size="18"/><span>{{ busy ? '请稍候…' : mode === 'login' ? '登录，开启对话' : '创建我的账号' }}</span><Icon v-if="!busy" name="ArrowRight" :size="19"/></button>
        </form>
        <div class="auth-divider"><span>先认识一下栖语</span></div><button class="button preview-button" @click="enterDemo" :disabled="busy"><Icon name="Compass" :size="18"/>浏览界面预览<Icon name="ArrowUpRight" :size="16"/></button><p class="preview-note">预览使用示例内容，不连接真实账号。</p>
      </div>
      <div class="auth-bottom"><Icon name="MessageCircle" :size="16"/><span>一段好的对话，从这里开始。</span></div>
    </section>
  </main>
</template>

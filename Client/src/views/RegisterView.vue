<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { registerApi } from '@/api/auth'

const router = useRouter()
const account = ref('')
const password = ref('')
const email = ref('')
const error = ref('')
const success = ref(false)
const loading = ref(false)

const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

function extractError(e: any): string {
  const d = e?.response?.data
  if (d?.message) return d.message
  if (d?.reason) return d.reason
  if (d?.msg) return d.msg
  if (typeof d === 'string') return d
  if (e?.message) return e.message
  return '网络错误，请稍后重试'
}

async function handleRegister() {
  if (!account.value || !password.value) { error.value = '请输入账号和密码'; return }
  if (!email.value) { error.value = '请输入邮箱'; return }
  if (!emailRegex.test(email.value)) { error.value = '邮箱格式不正确'; return }
  loading.value = true; error.value = ''
  try {
    const { data } = await registerApi({
      account: account.value,
      password: password.value,
      email: email.value,
    })
    if (data.code === 100) {
      success.value = true
      setTimeout(() => router.push('/login'), 1500)
    } else {
      error.value = data.message || '注册失败（错误码:' + data.code + '）'
    }
  } catch (e: any) {
    error.value = extractError(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1>注册账号</h1>
      <template v-if="success">
        <div class="success-msg">注册成功，即将跳转到登录页...</div>
      </template>
      <template v-else>
        <form @submit.prevent="handleRegister">
          <div class="form-group">
            <input v-model="account" placeholder="账号" autocomplete="username" />
          </div>
          <div class="form-group">
            <input v-model="password" type="password" placeholder="密码" autocomplete="new-password" />
          </div>
          <div class="form-group">
            <input v-model="email" type="email" placeholder="邮箱" autocomplete="email" />
          </div>
          <div v-if="error" class="error-msg">{{ error }}</div>
          <button class="btn-primary" type="submit" :disabled="loading">
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>
        <div class="login-footer">
          已有账号？<router-link to="/login">立即登录</router-link>
        </div>
      </template>
    </div>
  </div>
</template>

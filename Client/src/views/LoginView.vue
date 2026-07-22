<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const account = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

function extractError(e: any): string {
  const d = e?.response?.data
  if (d?.message) return d.message
  if (d?.reason) return d.reason
  if (d?.msg) return d.msg
  if (typeof d === 'string') return d
  if (e?.message) return e.message
  return '网络错误，请稍后重试'
}

async function handleLogin() {
  if (!account.value || !password.value) { error.value = '请输入账号和密码'; return }
  loading.value = true; error.value = ''
  try {
    const deviceId = 'web_' + Math.random().toString(36).slice(2, 10)
    const { data } = await loginApi({
      account: account.value,
      password: password.value,
      platform: 'web',
      device_id: deviceId,
    })
    if (data.code === 100) {
      auth.setAuth(String(data.uid), account.value, data.access_token, data.refresh_token)
      router.replace('/im')
    } else {
      error.value = data.message || '登录失败（错误码:' + data.code + '）'
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
      <h1>🐾Neko Chat🐾</h1>
      <form @submit.prevent="handleLogin">
        <div class="form-group">
          <input v-model="account" placeholder="账号" autocomplete="username" />
        </div>
        <div class="form-group">
          <input v-model="password" type="password" placeholder="密码" autocomplete="current-password" />
        </div>
        <div v-if="error" class="error-msg">{{ error }}</div>
        <button class="btn-primary" type="submit" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
      <div class="login-footer">
        还没有账号？<router-link to="/register">立即注册</router-link>
      </div>
    </div>
  </div>
</template>

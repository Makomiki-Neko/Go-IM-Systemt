import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfoResponse } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const uid = ref<string | null>(localStorage.getItem('uid') || null)
  const account = ref(localStorage.getItem('account') || '')
  const accessToken = ref(localStorage.getItem('access_token') || '')
  const refreshToken = ref(localStorage.getItem('refresh_token') || '')
  const userInfo = ref<UserInfoResponse | null>(null)

  const isLoggedIn = computed(() => !!accessToken.value)

  function setAuth(_uid: string, _account: string, _accessToken: string, _refreshToken: string) {
    uid.value = _uid
    account.value = _account
    accessToken.value = _accessToken
    refreshToken.value = _refreshToken
    localStorage.setItem('uid', _uid)
    localStorage.setItem('account', _account)
    localStorage.setItem('access_token', _accessToken)
    localStorage.setItem('refresh_token', _refreshToken)
  }

  function setUserInfo(info: UserInfoResponse) {
    userInfo.value = info
  }

  function setAccessToken(token: string) {
    accessToken.value = token
    localStorage.setItem('access_token', token)
  }

  function setPhoto(photo: string) {
    if (userInfo.value) {
      userInfo.value = { ...userInfo.value, photo }
    }
  }

  function logout() {
    uid.value = null
    account.value = ''
    accessToken.value = ''
    refreshToken.value = ''
    userInfo.value = null
    localStorage.clear()
  }

  return { uid, account, accessToken, refreshToken, userInfo, isLoggedIn, setAuth, setUserInfo, setAccessToken, setPhoto, logout }
})

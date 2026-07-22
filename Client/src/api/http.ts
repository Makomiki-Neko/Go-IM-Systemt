import axios from 'axios'

const http = axios.create({
  baseURL: '/api',
  timeout: 15000,
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((cb) => cb(token))
  refreshSubscribers = []
}

http.interceptors.response.use(
  (res) => res,
  async (error) => {
    const originalRequest = error.config
    const status = error.response?.status

    /* 非 401 或不需重试，直接拒绝 */
    if (status !== 401 || originalRequest._retry) {
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise((resolve) => {
        refreshSubscribers.push((token: string) => {
          originalRequest.headers.Authorization = `Bearer ${token}`
          resolve(http(originalRequest))
        })
      })
    }

    originalRequest._retry = true
    isRefreshing = true

    const refreshToken = localStorage.getItem('refresh_token')
    const uid = localStorage.getItem('uid')

    if (!refreshToken || !uid) {
      isRefreshing = false
      window.location.href = '/login'
      return Promise.reject(error)
    }

    const account = localStorage.getItem('account') || ''
    try {
      const accessToken = localStorage.getItem('access_token') || ''
      const { data } = await axios.post('/api/user/heart', {
        uid,
        account,
        refresh_token: refreshToken,
        device_id: 'web',
        platform: 'web',
      }, {
        headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
      })
      if (data.access_token) {
        localStorage.setItem('access_token', data.access_token)
        onTokenRefreshed(data.access_token)
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`
        return http(originalRequest)
      } else {
        /* heart 成功但不需要刷新，用原 token 重试 */
        const token = localStorage.getItem('access_token') || ''
        originalRequest.headers.Authorization = `Bearer ${token}`
        return http(originalRequest)
      }
    } catch {
      /* 刷新失败 → 说明 refresh_token 已过期或被踢下线 → 跳到登录页 */
      localStorage.clear()
      window.location.href = '/login'
      return Promise.reject(error)
    } finally {
      isRefreshing = false
    }
  },
)

export default http

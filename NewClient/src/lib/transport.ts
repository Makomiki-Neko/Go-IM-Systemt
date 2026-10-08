import { decode, object, str, num, type JsonObject } from './protocol'

export class APIError extends Error { constructor(message: string, public status = 0) { super(message) } }
export interface Credentials { uid: string; account: string; accessToken: string; refreshToken: string; device: string }
const sessionKey = 'qiyu.session.v1'
export function readSession(): Credentials | null {
  try { const c = object(decode(localStorage.getItem(sessionKey) || 'null')); return c.accessToken && c.uid ? c as unknown as Credentials : null }
  catch { return null }
}
let credentials = readSession()
export function saveSession(c: Credentials | null) {
  credentials = c
  if (c) localStorage.setItem(sessionKey, JSON.stringify(c)); else localStorage.removeItem(sessionKey)
}
let refreshPromise: Promise<void> | null = null
export async function refreshSession(): Promise<void> {
  if (!credentials) throw new APIError('请重新登录', 401)
  if (refreshPromise) return refreshPromise
  const original = credentials
  refreshPromise = (async () => {
    const r = await request('/user/heart', { uid: original.uid, refresh_token: original.refreshToken, platform: 'web', device_id: original.device }, 'POST', false)
    if (credentials !== original) return
    if (r.access_token) saveSession({ ...original, accessToken: str(r.access_token) })
  })().finally(() => { refreshPromise = null })
  return refreshPromise
}
export async function request(path: string, body?: unknown, method = 'POST', retry = true): Promise<JsonObject> {
  const headers: Record<string, string> = {}
  if (credentials) headers.Authorization = `Bearer ${credentials.accessToken}`
  if (body !== undefined && !(body instanceof FormData)) headers['Content-Type'] = 'application/json'
  const response = await fetch(`${import.meta.env.VITE_API_BASE || '/api'}${path}`, { method, headers, body: body === undefined ? undefined : body instanceof FormData ? body : JSON.stringify(body), signal: AbortSignal.timeout(15000) })
  const text = await response.text()
  let data: JsonObject = {}
  try { data = object(decode(text)) } catch { if (response.ok) throw new APIError('服务器返回了无法解析的数据') }
  if (response.status === 401 && retry && credentials) { await refreshSession(); return request(path, body, method, false) }
  if (!response.ok) throw new APIError(str(data.msg || data.message, text || `请求失败 (${response.status})`), response.status)
  const code = num(object(data.base).code ?? data.code)
  if (![0, 100, 200, 201].includes(code)) throw new APIError(str(data.msg || data.message || object(data.base).msg, `操作未成功 (${code})`), code)
  return data
}

export interface Packet { type: string; data: unknown; req_id?: number }
interface Pending { expected: string; resolve: (v: unknown) => void; reject: (e: Error) => void; timer: ReturnType<typeof setTimeout> }
export class Socket {
  private ws: WebSocket | null = null
  private stopped = true
  private timer?: ReturnType<typeof setTimeout>
  private heartbeat?: ReturnType<typeof setInterval>
  private pending = new Map<number, Pending>()
  private seq = Math.floor(Math.random() * 1e9)
  private attempts = 0
  onPacket: (packet: Packet) => void = () => {}
  onState: (state: string) => void = () => {}
  onOpen: () => void = () => {}
  get connected() { return this.ws?.readyState === WebSocket.OPEN }
  connect() { this.stopped = false; this.open() }
  private open() {
    if (this.stopped || !credentials || this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) return
    this.onState('connecting')
    const url = new URL(import.meta.env.VITE_WS_URL || '/ws', location.href)
    url.protocol = url.protocol === 'https:' || url.protocol === 'wss:' ? 'wss:' : 'ws:'
    url.searchParams.set('token', credentials.accessToken)
    const ws = new WebSocket(url)
    this.ws = ws
    ws.onopen = () => {
      if (this.stopped) return ws.close()
      this.attempts = 0; this.onState('online'); this.onOpen()
      this.heartbeat = setInterval(() => this.notify('heartBeat', {}), 25000)
    }
    ws.onmessage = e => {
      if (this.stopped || this.ws !== ws) return
      try {
        const raw = object(decode(str(e.data)))
        const packet: Packet = { type: str(raw.type), data: raw.data, req_id: num(raw.req_id ?? raw.reqId) }
        const p = this.pending.get(packet.req_id || 0)
        // Legacy MQ echoes request IDs to other recipients too. Match response type as well.
        if (p && (packet.type === p.expected || packet.type === 'Error')) {
          clearTimeout(p.timer); this.pending.delete(packet.req_id!)
          if (packet.type === 'Error') p.reject(new APIError(str(object(packet.data).message || object(packet.data).ErrorMsg, '操作失败')))
          else p.resolve(packet.data)
        }
        this.onPacket(packet)
      } catch (error) { console.warn('无法解析服务器消息', error) }
    }
    ws.onerror = () => ws.close()
    ws.onclose = () => {
      if (this.ws !== ws) return
      clearInterval(this.heartbeat)
      for (const p of this.pending.values()) { clearTimeout(p.timer); p.reject(new APIError('连接已断开，发送结果待确认')) }
      this.pending.clear(); this.ws = null
      if (this.stopped) return
      this.onState('offline')
      this.timer = setTimeout(() => this.open(), Math.min(15000, 1000 * 2 ** this.attempts++))
    }
  }
  send(type: string, payload: unknown, expected: string): Promise<unknown> {
    if (!this.connected) return Promise.reject(new APIError('连接尚未就绪，请稍后重试'))
    const reqId = ++this.seq
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => { this.pending.delete(reqId); reject(new APIError('响应超时，操作结果待确认')) }, 20000)
      this.pending.set(reqId, { expected, resolve, reject, timer })
      this.ws!.send(JSON.stringify({ type, reqId, payload }))
    })
  }
  notify(type: string, payload: unknown) { if (this.connected) this.ws!.send(JSON.stringify({ type, reqId: ++this.seq, payload })) }
  close() {
    this.stopped = true; clearTimeout(this.timer); clearInterval(this.heartbeat)
    const ws = this.ws; this.ws = null; ws?.close()
    for (const p of this.pending.values()) { clearTimeout(p.timer); p.reject(new APIError('连接已关闭')) }
    this.pending.clear(); this.onState('offline')
  }
}

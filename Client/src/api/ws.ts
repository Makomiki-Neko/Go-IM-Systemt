/**
 * WebSocket 客户端
 * 遵循 Project.MD 规范：
 * - 全局自增 ReqID（断连重置）
 * - 请求-响应模式（仅用于期望回应的请求）
 * - 火-忘模式（用于消息发送等服务器不直接回应的场景）
 * - 自动心跳
 * - 断线自动重连（使用最新 token）
 */

type WsMessageHandler = (data: any) => void

class WsClient {
  private ws: WebSocket | null = null
  private handlers = new Map<string, WsMessageHandler[]>()
  private pendingMap = new Map<number, { resolve: Function; reject: Function; timeout: ReturnType<typeof setTimeout> }>()

  /* 自增 ReqID */
  private _reqId = 0
  get nextReqId(): number {
    this._reqId += 1
    return this._reqId
  }
  resetReqId(): void { this._reqId = 0 }

  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private heartTimer: ReturnType<typeof setInterval> | null = null
  private _connected = false

  onConnected: (() => void) | null = null
  onDisconnected: (() => void) | null = null

  get connected(): boolean { return this._connected }

  /** 每次连接前从 localStorage 取最新 token */
  private getToken(): string {
    return localStorage.getItem('access_token') || ''
  }

  connect(): void {
    this.resetReqId()
    this._doConnect()
  }

  private _doConnect(): void {
    const token = this.getToken()
    if (!token) return
    this.ws = new WebSocket(`ws://localhost:8889/ws?token=${token}`)

    this.ws.onopen = () => {
      this._connected = true
      this._startHeartbeat()
      this.onConnected?.()
    }

    this.ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data)
        const { req_id, type } = msg

        /* 请求-响应匹配（响应通过 MQ→Hub→Client 推送回来） */
        if (req_id && this.pendingMap.has(req_id)) {
          const p = this.pendingMap.get(req_id)!
          clearTimeout(p.timeout)
          this.pendingMap.delete(req_id)
          p.resolve(msg)
          return
        }

        /* 按 type 分发推送 */
        this._emit(type || msg.type, msg)
      } catch { /* ignore */ }
    }

    this.ws.onclose = () => {
      this._connected = false
      this._stopHeartbeat()
      /* 清空 pending，但不 reject（服务端可能已处理） */
      this.pendingMap.forEach((p) => clearTimeout(p.timeout))
      this.pendingMap.clear()
      this.onDisconnected?.()
      this._scheduleReconnect()
    }

    this.ws.onerror = () => this.ws?.close()
  }

  private _scheduleReconnect(): void {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.resetReqId()
      this._doConnect()
    }, 3000)
  }

  /* WS 心跳：每 25s 发送 */
  private _startHeartbeat(): void {
    this.heartTimer = setInterval(() => {
      this._sendRaw({ type: 'heartBeat', reqId: this.nextReqId })
    }, 25000)
  }
  private _stopHeartbeat(): void {
    if (this.heartTimer) { clearInterval(this.heartTimer); this.heartTimer = null }
  }

  /** 发送并等待响应（用于 GetNewPrivateMsg / GetHistoryPrivateMsg） */
  send(type: string, payload: any, reqId?: number): Promise<any> {
    const id = reqId ?? this.nextReqId
    const sent = this._sendRaw({ type, reqId: id, payload })
    return new Promise((resolve, reject) => {
      if (!sent) {
        reject(new Error('WS not connected'))
        return
      }
      const timeout = setTimeout(() => {
        this.pendingMap.delete(id)
        reject(new Error('WS request timeout'))
      }, 20000)
      this.pendingMap.set(id, { resolve, reject, timeout })
    })
  }

  /** 发送但不等待响应（用于 SendPrivateMsg / ack / updateFile） */
  sendOnly(type: string, payload: any, reqId?: number): number | null {
    const id = reqId ?? this.nextReqId
    return this._sendRaw({ type, reqId: id, payload }) ? id : null
  }

  private _sendRaw(msg: any): boolean {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
      return true
    }
    return false
  }

  /* 事件监听 */
  on(type: string, handler: WsMessageHandler): void {
    if (!this.handlers.has(type)) this.handlers.set(type, [])
    this.handlers.get(type)!.push(handler)
  }
  off(type: string, handler: WsMessageHandler): void {
    const arr = this.handlers.get(type)
    if (arr) {
      const idx = arr.indexOf(handler)
      if (idx >= 0) arr.splice(idx, 1)
    }
  }
  private _emit(type: string, data: any): void {
    this.handlers.get(type)?.forEach((h) => h(data))
  }

  disconnect(): void {
    this._stopHeartbeat()
    if (this.reconnectTimer) { clearTimeout(this.reconnectTimer); this.reconnectTimer = null }
    this.ws?.close()
    this.ws = null
    this._connected = false
    this._reqId = 0
    this.pendingMap.forEach((p) => { clearTimeout(p.timeout); p.reject(new Error('WS disconnected')) })
    this.pendingMap.clear()
  }
}

export const wsClient = new WsClient()

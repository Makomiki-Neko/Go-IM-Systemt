import { parse, isLosslessNumber } from 'lossless-json'

export type JsonObject = Record<string, unknown>
export const object = (v: unknown): JsonObject => v !== null && typeof v === 'object' && !Array.isArray(v) ? v as JsonObject : {}
export const array = (v: unknown): unknown[] => Array.isArray(v) ? v : []
export const str = (v: unknown, fallback = ''): string => v === null || v === undefined ? fallback : String(v)
export const num = (v: unknown): number => Number(v) || 0

/** Unsafe integer tokens become strings before JS can round AI history IDs. */
export function decode(text: string): unknown {
  return parse(text, (_key, value: unknown) => {
    if (!isLosslessNumber(value)) return value
    const n = Number(value.value)
    return Number.isInteger(n) && !Number.isSafeInteger(n) ? value.value : n
  })
}
export const idCompare = (a: string, b: string) => BigInt(a || '0') < BigInt(b || '0') ? -1 : BigInt(a || '0') > BigInt(b || '0') ? 1 : 0
export const maxID = (a: string, b: string) => idCompare(a, b) >= 0 ? a : b
export function safeURL(path: string, base = ''): string {
  if (!path) return ''
  try {
    const url = new URL(/^https?:\/\//.test(path) ? path : `${base.replace(/\/$/, '')}/${path.replace(/^\//, '')}`, location.origin)
    return ['http:', 'https:'].includes(url.protocol) ? url.href : ''
  } catch { return '' }
}

export type Kind = 'private' | 'group' | 'ai'
export type SendState = 'sending' | 'processing' | 'queued' | 'running' | 'succeeded' | 'failed' | 'unknown'
export interface Message {
  id: string; clientId?: string; sender: string; content: string; type: number; time: number
  reasoning?: string; attempt?: number;
  seq?: string; turn?: string; taskId?: string; state?: SendState; error?: string; mentions?: string[]; version?: string
}
export interface Conversation {
  key: string; id: string; kind: Kind; name: string; avatar: string; subtitle: string; online: boolean
  unread: number; preview: string; updated: number; version: string; visibleAfter: string; readSeq: string; lastSeq: string
}
export interface Friend { id: string; name: string; remark: string; avatar: string; signature: string; online: boolean }
export function friendFrom(v: unknown): Friend {
  const f = object(v)
  return { id: str(f.id), name: str(f.friend_name), remark: str(f.remark), avatar: str(f.avatar), signature: str(f.signature), online: !!f.online_statu }
}
export function messageFrom(v: unknown, kind: Kind): Message {
  const m = object(v)
  if (kind === 'ai') return { id: str(m.MsgID), sender: m.IsAiMsg ? 'ai' : str(m.UserID), content: str(m.Content), reasoning: str(m.Reasoning), type: num(m.MsgType) || 1, state: m.IsAiMsg ? 'succeeded' : undefined, time: Date.parse(str(m.UpdatedAt)) || Date.now(), turn: str(m.TurnSeq) }
  return { id: str(m.msg_id), clientId: m.client_msg_id ? str(m.client_msg_id) : undefined, sender: str(m.from_user_id), content: str(m.content), type: num(m.msg_type), time: num(m.send_time) * (kind === 'private' ? 1000 : 1), seq: m.seq ? str(m.seq) : undefined, mentions: array(m.mentions).map(v => str(v)), state: 'succeeded' }
}
export function mergeMessages(current: Message[], incoming: Message[]): Message[] {
  const out = [...current]
  for (const m of incoming) {
    const index = out.findIndex(x => x.id === m.id || (!!m.clientId && x.clientId === m.clientId && x.sender === m.sender))
    if (index < 0) out.push(m)
    else out[index] = { ...out[index], ...m, state: m.state ?? out[index].state }
  }
  return out.sort((a, b) => {
    if (a.seq && b.seq) return idCompare(a.seq, b.seq)
    if (a.turn && b.turn) return idCompare(a.turn, b.turn) || (a.sender === 'ai' ? 1 : 0) - (b.sender === 'ai' ? 1 : 0)
    if (/^\d+$/.test(a.id) && /^\d+$/.test(b.id)) return idCompare(a.id, b.id)
    return a.time - b.time
  })
}

/** Different browser tabs use random safe integers; private WS requires a number. */
export function clientID(): string {
  const bytes = crypto.getRandomValues(new Uint32Array(2))
  return String((BigInt(bytes[0] & 0x1fffff) << 32n) | BigInt(bytes[1]) || 1n)
}

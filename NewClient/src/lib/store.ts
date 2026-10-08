import { AIPlayback } from './aiPlayback'
import { reactive, computed } from 'vue'
import { request, refreshSession, readSession, saveSession, Socket, APIError, type Credentials, type Packet } from './transport'
import { array, object, str, num, decode, clientID, maxID, idCompare, friendFrom, messageFrom, mergeMessages, type Conversation, type Friend, type Message, type JsonObject, type Kind } from './protocol'
import { demoContent, demoFriends } from './demo'

const socket = new Socket()
export const agent = reactive({ id: '0', name: '', avatar: '', description: '' })
const removedAI = new Set<string>()
const aiPlayback = new Map<string, AIPlayback>()
function stopPlayback(key:string) { aiPlayback.get(key)?.stop(); aiPlayback.delete(key) }
function aiPlaceholder(c:Conversation, cid:string):Message {
  let m=(state.messages[bucket(c)] || []).find(m=>m.sender==='ai' && m.clientId===cid)
  if (!m) { add(c,[{id:`ai-pending-${cid}`,clientId:cid,sender:'ai',content:'',reasoning:'',type:1,time:Date.now()+1,state:'queued'}]);m=state.messages[bucket(c)].find(m=>m.sender==='ai' && m.clientId===cid)! }
  return m
}
function streamAI(c:Conversation,d:JsonObject) {
  const cid=str(d.client_msg_id), key=`${c.key}:${cid}`, m=aiPlaceholder(c,cid)
  if (m.state==='succeeded' || m.state==='failed') return
  const attempt=num(d.attempt)
  if (m.attempt!==attempt) {stopPlayback(key);m.content='';m.reasoning='';m.attempt=attempt}
  m.state='running';m.taskId=str(d.task_id);m.turn=str(d.turn_seq)
  if (!aiPlayback.has(key)) aiPlayback.set(key,new AIPlayback((content,reasoning)=>{
    const target=(state.messages[bucket(c)]||[]).find(m=>m.sender==='ai' && m.clientId===cid)
    if (target && target.state==='running') {target.content=content;target.reasoning=reasoning}
  }))
  aiPlayback.get(key)!.push(num(d.seq),str(d.content),str(d.reasoning_content))
}

let aiRevision = 0
export const state = reactive({
  session: readSession() as Credentials | null, demo: false, connection: 'offline',
  conversations: [] as Conversation[], friends: [] as Friend[], applications: [] as Friend[],
  messages: {} as Record<string, Message[]>, cursors: {} as Record<string, string>,
  active: '', tab: 'chats' as 'chats' | 'contacts' | 'ai', filter: 'all', search: '',
  profile: { name: '', photo: '', signature: '', gender: '', birthday: 0 },
  loading: false, syncing: {} as Record<string, boolean>, noHistory: {} as Record<string, boolean>,
  toasts: [] as { id: number; text: string; error: boolean }[], error: '',
})
export const current = computed(() => state.conversations.find(c => c.key === state.active))
export function bucket(c: Conversation) { return c.key + (c.kind === 'group' ? `@${c.version}` : '') }
export const currentMessages = computed(() => current.value ? state.messages[bucket(current.value)] || [] : [])
let heart: ReturnType<typeof setInterval> | undefined
let poll: ReturnType<typeof setInterval> | undefined
let generation = 0
let heartBusy = false
let polling = false
let saveTimer: ReturnType<typeof setTimeout> | undefined
let cacheEnabled = true

export function toast(text: string, error = false) {
  const id = Date.now() + Math.random(); state.toasts.push({ id, text, error })
  setTimeout(() => { state.toasts = state.toasts.filter(t => t.id !== id) }, 5000)
}
export function explain(error: unknown) { return error instanceof Error ? error.message : '操作未完成，请稍后重试' }
export async function run(action: () => Promise<unknown>, success = '') {
  try { await action(); if (success) toast(success); return true }
  catch (error) { toast(explain(error), true); return false }
}
function persist() {
  clearTimeout(saveTimer)
  const owner = state.session?.uid
  if (!owner || state.demo || !cacheEnabled) return
  saveTimer = setTimeout(() => {
    if (state.session?.uid !== owner || state.demo || !cacheEnabled) return
    try { localStorage.setItem(`qiyu.cache.v1:${owner}`, JSON.stringify({ conversations: state.conversations, messages: state.messages, cursors: state.cursors })) }
    catch { toast('本地存储空间不足，本次记录仅在当前页面保留', true) }
  }, 250)
}
export function clearCache() {
  clearTimeout(saveTimer); cacheEnabled = false
  localStorage.removeItem(`qiyu.cache.v1:${state.session?.uid}`)
  const prefix = `qiyu.draft:${state.session?.uid}:`
  for (const key of Object.keys(sessionStorage)) if (key.startsWith(prefix)) sessionStorage.removeItem(key)
}
function loadCache() {
  try {
    const cache = object(decode(localStorage.getItem(`qiyu.cache.v1:${state.session?.uid}`) || '{}'))
    state.conversations = array(cache.conversations) as Conversation[]
    state.messages = object(cache.messages) as Record<string, Message[]>
    state.cursors = object(cache.cursors) as Record<string, string>
    // Unacknowledged requests are uncertain after a reload, never silently resend them.
    for (const list of Object.values(state.messages)) for (const m of list) if (m.state === 'sending') m.state = 'unknown'
  } catch { state.conversations = []; state.messages = {}; state.cursors = {} }
}
export async function login(account: string, password: string) {
  const device = localStorage.getItem('qiyu.device') || crypto.randomUUID()
  localStorage.setItem('qiyu.device', device)
  const data = await request('/user/login', { account, password, device_id: device, platform: 'web' }, 'POST', false)
  if (!data.access_token || !data.uid) throw new Error('登录未成功，请检查账号与密码')
  const session = { uid: str(data.uid), account, accessToken: str(data.access_token), refreshToken: str(data.refresh_token), device }
  saveSession(session); state.session = session; state.demo = false
  await start()
}
export async function register(account: string, password: string, email: string) {
  await request('/user/register', { account, password, email }, 'POST', false)
}
export async function start() {
  if (!state.session || state.demo) return
  generation++; cacheEnabled = true; state.error = ''; loadCache()
  socket.onState = status => { state.connection = status }
  socket.onPacket = receive
  socket.onOpen = () => { if (current.value) void run(() => sync(current.value!)) }
  socket.connect()
  await reload()
  heart = setInterval(() => { void keepAlive() }, 25000)
  poll = setInterval(() => { void periodic() }, 12000)
  void keepAlive()
}
async function keepAlive() {
  if (heartBusy || !state.session || state.demo) return
  const epoch = generation
  heartBusy = true
  try { await refreshSession(); if (epoch === generation) state.error = '' }
  catch (e) {
    if (epoch !== generation) return
    if (e instanceof APIError && [401, 403, 412].includes(e.status)) { logout(); toast('登录已失效，请重新登录', true) }
    else state.error = '暂时无法连接服务，正在重试'
  } finally { heartBusy = false }
}
async function periodic() {
  if (polling || state.demo || !state.session) return
  polling = true
  try {
    if (current.value && socket.connected) await sync(current.value)
    await refreshUnread()
    if (socket.connected) {
      for (const c of state.conversations.filter(c => c.kind === 'ai')) {
        for (const m of (state.messages[bucket(c)] || []).filter(m => m.sender !== 'ai' && m.taskId && ['queued', 'running', 'processing', 'unknown'].includes(m.state || '')).slice(-5)) {
          const d = object(await socket.send('ai.GetAiTask', { session_id: c.id, task_id: m.taskId }, 'ai.Task'))
          updateAck(c, m.clientId || '', d)
          if (d.state === 'succeeded') await sync(c)
        }
      }
    }
  } catch (e) { state.error = explain(e) }
  finally { polling = false }
}
export function logout() {
	removedAI.clear(); aiRevision++; Object.assign(agent,{id:'0',name:'',avatar:'',description:''})
  for (const key of aiPlayback.keys()) stopPlayback(key)
  generation++; socket.close(); clearInterval(heart); clearInterval(poll); clearTimeout(saveTimer)
  saveSession(null); state.session = null; state.demo = false; state.active = ''; state.conversations = []; state.messages = {}; state.cursors = {}; state.friends = []; state.applications = []; state.syncing = {}; state.noHistory = {}; state.error = ''; state.search = ''; state.tab = 'chats'
}
export function enterDemo() {
  state.demo = true; state.session = { uid: '101', account: 'wanderer', accessToken: '', refreshToken: '', device: '' }
  state.profile = { name: '周予', photo: '', signature: '连接有趣的人，收藏平凡的日常。', gender: '', birthday: 0 }
  const demo = demoContent(); state.conversations = demo.conversations; state.messages = demo.messages
  state.friends = [...demoFriends]; state.connection = 'demo'; state.active = 'group:301'; state.error = ''
}
export function conversation(id: string, name: string, kind: Kind): Conversation {
  return { key: `${kind}:${id}`, id, name, kind, avatar: '', subtitle: '', online: false, unread: 0, preview: '', updated: 0, version: '0', visibleAfter: '0', readSeq: '0', lastSeq: '0' }
}
export async function reload() {
  if (state.demo) return
  state.loading = true
  const epoch = generation
  const results = await Promise.allSettled([
    friendPages('list'), friendPages('apply_list'),
    request(`/user/info?account=${encodeURIComponent(state.session?.account || '')}`, undefined, 'GET'), request('/group/sessions', {}),
  ])
  if (epoch !== generation) return
  const [friends, apps, profile, groups] = results
  if (friends.status === 'fulfilled') {
    state.friends = array(friends.value.list).map(friendFrom)
    for (const f of state.friends) {
      let c = state.conversations.find(c => c.key === `private:${f.id}`)
      if (!c) { c = conversation(f.id, f.remark || f.name, 'private'); state.conversations.push(c) }
      Object.assign(c, { name: f.remark || f.name, avatar: f.avatar, online: f.online, subtitle: f.signature })
    }
    const ids = new Set(state.friends.map(f => f.id))
    state.conversations = state.conversations.filter(c => c.kind !== 'private' || ids.has(c.id))
  }
  if (apps.status === 'fulfilled') state.applications = array(apps.value.list).map(friendFrom)
  if (profile.status === 'fulfilled') Object.assign(state.profile, profile.value)
  if (groups.status === 'fulfilled') applyGroups(array(groups.value.sessions))
  state.error = results.some(r => r.status === 'rejected') ? '部分资料加载失败，可点击刷新重试' : ''
  state.loading = false; persist()
  await refreshUnread().catch(e => { state.error = explain(e) })
  await reloadAI().catch(e => { state.error = explain(e) })
}
async function refreshUnread() {
  const epoch = generation
  const d = await request('/chat/GetUnreadPrivateMsgNumber', undefined, 'GET')
  if (epoch !== generation) return
  for (const c of state.conversations.filter(c => c.kind === 'private')) c.unread = 0
  for (const u of array(d.list).map(object)) {
    const c = state.conversations.find(c => c.key === `private:${str(u.id)}`); if (c) c.unread = num(u.count)
  }
}
function applyGroups(rows: unknown[]) {
  const ids = new Set<string>()
  for (const r of rows.map(object)) {
    const id = str(r.group_id); ids.add(id)
    let c = state.conversations.find(c => c.key === `group:${id}`)
    if (!c) { c = conversation(id, str(r.name), 'group'); state.conversations.push(c) }
    Object.assign(c, { name: str(r.name), version: str(r.membership_version), lastSeq: str(r.last_seq), readSeq: str(r.last_read_seq), visibleAfter: str(r.visible_after_seq), unread: r.has_unread ? 1 : 0 })
  }
  state.conversations = state.conversations.filter(c => c.kind !== 'group' || ids.has(c.id))
  if (state.active && state.active !== 'ai-settings' && !current.value) state.active = ''
  persist()
  if (current.value?.kind === 'group') void run(() => sync(current.value!))
}
function add(c: Conversation, rows: Message[]) {
  if (c.kind === 'ai' && removedAI.has(c.id)) return
  const key = bucket(c)
  if (c.kind==='ai') for(const row of rows) if(row.sender==='ai' && row.turn && /^\d+$/.test(row.id)) {
    const pending=(state.messages[key]||[]).find(m=>m.sender==='ai' && m.turn===row.turn && m.clientId)
    if(pending) {row.clientId=pending.clientId;stopPlayback(`${c.key}:${pending.clientId}`)}
  }
  state.messages[key] = mergeMessages(state.messages[key] || [], rows)
  const latest = state.messages[key].at(-1)
  if (latest) { c.preview = latest.type === 1 ? latest.content : latest.type === 2 ? '[图片]' : '[附件]'; c.updated = latest.time }
  persist()
}
function receive(packet: Packet) {
  if (packet.type === 'chat.groupSessions') { applyGroups(array(packet.data)); return }
  const d = object(packet.data)
  if (packet.type === 'chat.groupMsg') {
    const c = state.conversations.find(c => c.key === `group:${str(d.group_id)}`)
    if (!c || idCompare(str(d.seq), c.visibleAfter) <= 0) return
    add(c, [{ ...messageFrom(d, 'group'), version: c.version }]); c.lastSeq = maxID(c.lastSeq, str(d.seq))
    if (str(d.from_user_id) !== state.session?.uid && d.notify) c.unread = 1
    // Do not advance the contiguous cursor from broadcasts, which can arrive out of order.
    if (state.active === c.key) void run(() => sync(c))
  } else if (packet.type === 'chat.privateMsg') {
    const id = str(d.from_user_id) === state.session?.uid ? str(d.to_user_id) : str(d.from_user_id)
    const c = state.conversations.find(c => c.key === `private:${id}`)
    if (c) { add(c, [messageFrom(d, 'private')]); if (str(d.from_user_id) !== state.session?.uid) c.unread++ }
  } else if (packet.type === 'event.friendApply') { void reload() }
  else if (packet.type.startsWith('ai.LlmResponse.')) {
    const c = state.conversations.find(c => c.key === `ai:${str(d.session_id)}`)
    if (!c) return
    if (packet.type==='ai.LlmResponse.Start' || packet.type==='ai.LlmResponse.Delta') {streamAI(c,d);return}
    const cid=str(d.client_msg_id)
    stopPlayback(`${c.key}:${cid}`)
    updateAck(c, cid, { ...d, msg_id: undefined })
    const placeholder=aiPlaceholder(c,cid)
    if (d.state === 'succeeded') {
      state.messages[bucket(c)]=state.messages[bucket(c)].filter(m=>m.id!==str(d.msg_id) || m===placeholder)
      add(c, [{ id: str(d.msg_id), clientId:cid, sender: 'ai', content: str(d.content), reasoning:str(d.reasoning_content), state:'succeeded', type: 1, time: placeholder.time, turn: str(d.turn_seq) }]); if (state.active !== c.key) c.unread = 1
    } else {placeholder.state='failed';placeholder.error=str(d.error_message);persist()}
  }
}
export async function select(c: Conversation) {
  state.active = c.key; state.tab = c.kind === 'ai' ? 'ai' : 'chats'
  if (state.demo) { c.unread = 0; return }
  await run(() => sync(c))
}
export async function sync(c: Conversation) {
  const key = bucket(c), epoch = generation
  if (state.demo || state.syncing[key] || !socket.connected) return
  state.syncing[key] = true
  try {
    if (c.kind !== 'group' && !state.cursors[key]) {
      const rows = array(await socket.send(c.kind === 'ai' ? 'ai.GetHistoryAiMsg' : 'chat.GetHistoryPrivateMsg', c.kind === 'ai' ? { session_id: c.id, start_msg_id: '0', limit: 50 } : { from_user_id: c.id, start_msg_id: '18446744073709551615', limit: 50 }, c.kind === 'ai' ? 'ai.HistoryMsgBlock' : 'chat.privateHistoryMsgBlock')).map(r => messageFrom(r, c.kind))
      if (epoch !== generation) return
      add(c, rows); state.noHistory[key] = rows.length < 50
      state.cursors[key] = rows.reduce((id, m) => maxID(id, m.id), '0')
    }
    let more = true
    while (more && epoch === generation && bucket(c) === key) {
      const cursor = state.cursors[key] || c.visibleAfter
      const raw = await socket.send(c.kind === 'group' ? 'chat.GetNewGroupMsg' : c.kind === 'ai' ? 'ai.GetNewAiMsg' : 'chat.GetNewPrivateMsg', c.kind === 'group' ? { group_id: c.id, membership_version: c.version, cursor_seq: cursor, limit: 200 } : c.kind === 'ai' ? { session_id: c.id, start_msg_id: cursor, limit: 200 } : { from_user_id: c.id, start_msg_id: cursor, limit: 200 }, c.kind === 'group' ? 'chat.groupNewMsgBlock' : c.kind === 'ai' ? 'ai.NewMsgBlock' : 'chat.privateUnreceiveMsgBlock')
      if (epoch !== generation || bucket(c) !== key) return
      const d = object(raw)
      const rows = array(c.kind === 'group' ? d.messages : raw).map(r => messageFrom(r, c.kind))
      add(c, rows)
      const next = c.kind === 'group' ? str(d.next_cursor, cursor) : rows.reduce((id, m) => maxID(id, m.id), cursor)
      state.cursors[key] = maxID(cursor, next)
      more = c.kind === 'group' ? !!d.has_more : rows.length === 200
      if (more && idCompare(next, cursor) <= 0) throw new Error('消息分页没有前进，请刷新后重试')
      persist()
    }
  } finally { if (epoch === generation) state.syncing[key] = false }
}
export async function history(c: Conversation) {
  if (state.demo) { state.noHistory[bucket(c)] = true; return }
  const key = bucket(c), epoch = generation
  const rows = (state.messages[key] || []).filter(m => /^\d+$/.test(m.id))
  const earliest = rows[0]
  const raw = await socket.send(c.kind === 'group' ? 'chat.GetHistoryGroupMsg' : c.kind === 'ai' ? 'ai.GetHistoryAiMsg' : 'chat.GetHistoryPrivateMsg', c.kind === 'group' ? { group_id: c.id, membership_version: c.version, cursor_seq: earliest?.seq || '0', limit: 50 } : c.kind === 'ai' ? { session_id: c.id, start_msg_id: earliest?.id || '0', limit: 50 } : { from_user_id: c.id, start_msg_id: earliest?.id || '18446744073709551615', limit: 50 }, c.kind === 'group' ? 'chat.groupHistoryMsgBlock' : c.kind === 'ai' ? 'ai.HistoryMsgBlock' : 'chat.privateHistoryMsgBlock')
  if (epoch !== generation || bucket(c) !== key) return
  const list = array(c.kind === 'group' ? object(raw).messages : raw).map(r => messageFrom(r, c.kind))
  add(c, list); state.noHistory[key] = c.kind === 'group' ? !object(raw).has_more : list.length < 50
}
function updateAck(c: Conversation, client: string, d: JsonObject) {
  const m = (state.messages[bucket(c)] || []).find(m => m.clientId === client && m.sender !== 'ai')
  if (!m) return
  // A late acceptance ACK must not overwrite a terminal event.
  const keepTerminal = ['succeeded', 'failed'].includes(m.state || '') && ['queued', 'running', 'processing'].includes(str(d.state))
  if (d.msg_id && str(d.msg_id) !== '0') m.id = str(d.msg_id)
  if (d.turn_seq) m.turn = str(d.turn_seq)
  if (d.seq) m.seq = str(d.seq)
  if (d.task_id) m.taskId = str(d.task_id)
  if (d.send_time) m.time = num(d.send_time) * (c.kind === 'private' ? 1000 : 1)
  if (c.kind==='ai') {const reply=(state.messages[bucket(c)]||[]).find(x=>x.sender==='ai' && x.clientId===client);if(reply)reply.time=Math.max(reply.time,m.time+1)}
  if (!keepTerminal) { m.state = (str(d.state) || 'succeeded') as Message['state']; m.error = str(d.error_message) }
  state.messages[bucket(c)] = mergeMessages([], state.messages[bucket(c)] || [])
  persist()
}
export async function send(c: Conversation, content: string, type = 1, mentions: string[] = [], retry?: Message) {
  if (!content.trim()) return
  if (!state.demo && !socket.connected) throw new Error('连接尚未就绪，内容已保留')
  const cid = retry?.clientId || clientID()
  const key = bucket(c), epoch = generation
  const m: Message = retry || { id: `local-${cid}`, clientId: cid, sender: state.session!.uid, content, type, time: Date.now(), mentions, version: c.version, state: 'sending' }
  m.state = 'sending'; m.error = ''; add(c, [m])
  if (c.kind==='ai' && !state.demo) aiPlaceholder(c,cid)
  if (state.demo) { m.id = clientID(); m.state = 'succeeded'; add(c, [m]); return }
  try {
    const d = object(await socket.send(c.kind === 'group' ? 'chat.SendGroupMsg' : c.kind === 'ai' ? 'ai.SendMsgToAi' : 'chat.SendPrivateMsg', { ...(c.kind === 'group' ? { group_id: c.id, mention_user_ids: mentions } : c.kind === 'ai' ? { session_id: c.id } : { to_user_id: c.id }), client_msg_id: c.kind === 'private' ? Number(cid) : cid, msg_type: type, content }, c.kind === 'group' ? 'ack.GroupMsg' : c.kind === 'ai' ? 'ack.AiMsg' : 'ack.Msg'))
    if (epoch !== generation || bucket(c) !== key) return
    if (!d.msg_id || str(d.msg_id) === '0') {
      if (!['failed', 'processing'].includes(str(d.state))) throw new Error('服务未返回有效消息 ID，请先刷新记录确认结果')
    }
    updateAck(c, cid, d)
    if(c.kind==='ai') {const reply=aiPlaceholder(c,cid);if(d.task_id)reply.taskId=str(d.task_id);if(d.state==='failed'){stopPlayback(`${c.key}:${cid}`);reply.state='failed';reply.error=str(d.error_message)}}
    if (c.kind !== 'ai') await sync(c)
  } catch (e) {
    if (epoch !== generation || bucket(c) !== key) return
    const existing = (state.messages[key] || []).find(x => x.clientId === cid)
    if (existing && existing.state !== 'succeeded' && existing.state !== 'failed') { existing.state = 'unknown'; existing.error = explain(e); if(c.kind==='ai'){const reply=aiPlaceholder(c,cid);if(reply.state==='queued'){reply.state='unknown';reply.error=explain(e)}} persist() }
  }
}
export async function markRead(c: Conversation) {
  if (document.visibilityState !== 'visible' || !document.hasFocus()) return
  if (state.demo || c.kind === 'ai') { c.unread = 0; return }
  if (!socket.connected || state.syncing[bucket(c)]) return
  if (c.kind === 'group') {
    const seq = state.cursors[bucket(c)] || c.visibleAfter
    if (idCompare(seq, c.readSeq) <= 0) return
    await socket.send('ack.GroupMsgRead', { group_id: c.id, membership_version: c.version, seq }, 'ack.GroupRead')
    c.readSeq = maxID(c.readSeq, seq); c.unread = 0
  } else {
    const id = (state.messages[bucket(c)] || []).filter(m => /^\d+$/.test(m.id)).at(-1)?.id
    if (id) { socket.notify('ack.PrivateMsgRead', { target_id: c.id, msg_id: id }); c.unread = 0 }
  }
  persist()
}
export async function createAI() {
  if (!state.demo) await loadAgent()
  const id = state.demo ? clientID() : str(object(await socket.send('ai.CreateSession', { agent_id: agent.id }, 'ai.Session')).session_id)
  if (!id) throw new Error('未取得会话 ID')
  aiRevision++
  const c = conversation(id, agent.name ? `${agent.name} · ${id.slice(-4)}` : '新对话', 'ai'); c.avatar = agent.avatar; c.updated = Date.now(); state.conversations.unshift(c); persist(); await select(c)
}
export async function loadAgent() {
  if (state.demo) return
  const epoch = generation
  const r = await request('/ai/agent/get', {})
  if (epoch !== generation) return
  Object.assign(agent, {id:str(r.agent_id,'0'), name:str(r.agent_name), avatar:str(r.agent_avatar), description:str(r.agent_prompt)})
}
export async function saveAgent(name: string, description: string, avatar: string) {
  if (state.demo) throw new Error('请登录后设置智能体')
  await request('/ai/agent/set', { agent_id: agent.id, agent_name: name, agent_prompt: description, agent_avatar: avatar })
  await loadAgent()
}
export async function reloadAI() {
  if (state.demo) return
  const epoch = generation, revision = aiRevision, rows: JsonObject[] = []
  for (let page = 1; ; page++) {
    const r = await request('/ai/sessions', { page: { page, size: 100 } })
    const chunk = array(r.list).map(object); rows.push(...chunk)
    if (!chunk.length || rows.length >= num(object(r.page).total)) break
  }
  if (epoch !== generation || revision !== aiRevision) return
  const ids = new Set(rows.map(r => str(r.session_id)))
  for (const c of state.conversations.filter(c => c.kind === 'ai' && !ids.has(c.id))) dropAI(c)
  for (const r of rows) {
    const id = str(r.session_id); if (removedAI.has(id)) continue
    let c = state.conversations.find(c => c.key === `ai:${id}`)
    if (!c) { c = conversation(id, str(r.agent_name, 'AI 对话') + ` · ${id.slice(-4)}`, 'ai'); state.conversations.push(c) }
    c.avatar = str(r.agent_avatar); if (!c.updated) c.updated = Math.max(0,num(r.created_at))
  }
  persist()
}
function dropAI(c: Conversation) {
  for (const key of aiPlayback.keys()) if(key.startsWith(`${c.key}:`)) stopPlayback(key)
  removedAI.add(c.id)
  state.conversations = state.conversations.filter(x => x.key !== c.key)
  delete state.messages[bucket(c)]; delete state.cursors[bucket(c)]; delete state.noHistory[bucket(c)]
  sessionStorage.removeItem(`qiyu.draft:${state.session?.uid}:${bucket(c)}`)
  if (state.active === c.key) state.active = ''
}
export async function deleteAI(c: Conversation) {
  if (!state.demo) await request('/ai/session/delete', {session_id:c.id})
  aiRevision++; dropAI(c); persist()
}
export async function restoreAI(id: string) {
  if (!/^\d+$/.test(id)) throw new Error('请输入正确的会话 ID')
  const c = state.conversations.find(c => c.key === `ai:${id}`) || conversation(id, 'AI 对话', 'ai')
  // History request validates ownership before keeping an imported session.
  if (!state.demo) await socket.send('ai.GetHistoryAiMsg', { session_id: id, start_msg_id: '0', limit: 1 }, 'ai.HistoryMsgBlock')
  if (!state.conversations.includes(c)) state.conversations.unshift(c)
  persist(); await select(c)
}
export async function upload(c: Conversation, file: File) {
  if (state.demo) throw new Error('界面预览不上传文件，请登录后使用')
  if (file.size <= 0 || file.size > 50 * 1024 * 1024) throw new Error('请选择 50 MB 以内的非空文件')
  if (c.kind === 'ai') throw new Error('AI 暂时仅支持文字')
  const category = file.type.startsWith('image/') ? 'Picture' : file.type.startsWith('video/') ? 'Video' : file.type.startsWith('audio/') ? 'Audio' : 'File'
  const d = object(await socket.send(`updateFile.${category}`, { file_id: crypto.randomUUID(), file_name: file.name, file_size: Math.ceil(file.size / 1024), file_type: file.type }, 'file.UpdateUrl'))
  const url = new URL(str(d.url)); if (!['http:', 'https:'].includes(url.protocol)) throw new Error('无效的上传地址')
  const response = await fetch(url, { method: 'PUT', body: file, signal: AbortSignal.timeout(120000) })
  if (!response.ok) throw new Error(`上传失败 (${response.status})，尚未发送消息`)
  await send(c, str(d.fileId), category === 'Picture' ? 2 : category === 'Video' ? 4 : category === 'Audio' ? 3 : 5)
}
export async function groupAction(action: string, body: JsonObject = {}) {
  if (state.demo) throw new Error('界面预览不修改群资料，请登录后操作')
  return request(`/group/${action}`, body)
}
export async function groupList(action: 'members' | 'requests', groupID: string): Promise<JsonObject[]> {
  const rows: JsonObject[] = []
  for (let page = 1; ; page++) {
    const d = await groupAction(action, { group_id: groupID, page: { page, size: 100 } })
    const chunk = array(d.list).map(object); rows.push(...chunk)
    if (!chunk.length || rows.length >= num(object(d.page).total)) return rows
  }
}
async function friendPages(action: string): Promise<JsonObject> {
  const rows: unknown[] = []
  for (let page = 1; ; page++) {
    const d = await request(`/friend/${action}?page=${page}&size=200`, {})
    const chunk = array(d.list); rows.push(...chunk)
    if (chunk.length < 200) return { ...d, list: rows }
  }
}
export async function friendAction(action: string, body: JsonObject) {
  if (state.demo) throw new Error('界面预览不修改好友关系，请登录后操作')
  return request(`/friend/${action}`, body)
}
export async function updateProfile(data: typeof state.profile, file?: File) {
  if (state.demo) { Object.assign(state.profile, data); return }
  if (file) { const form = new FormData(); form.append('photo', file); const r = await request('/user/update/avatar', form); data.photo = str(r.photo_id) }
  await request('/user/update/info', data); Object.assign(state.profile, data)
}

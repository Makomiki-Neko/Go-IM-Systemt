import { test, expect, type Page } from '@playwright/test'

test('Preview: desktop navigation, messaging, details, profile and logout', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', e => errors.push(e.message))
  await page.goto('/')
  await expect(page.getByRole('heading', { name: '很高兴，又见面了' })).toBeVisible()
  await page.screenshot({ path: 'test-results/auth-desktop.png', fullPage: true, animations: 'disabled' })
  await page.getByRole('button', { name: '浏览界面预览' }).click()
  await expect(page.getByRole('heading', { name: '设计散步小组' })).toBeVisible()
  await page.screenshot({ path: 'test-results/chat-desktop.png', fullPage: true, animations: 'disabled' })
  await page.getByRole('textbox', { name: '消息内容' }).fill('新的界面，慢慢聊。')
  await page.getByRole('button', { name: '发送消息', exact: true }).click()
  await expect(page.getByText('新的界面，慢慢聊。', { exact: true }).last()).toBeVisible()
  await page.getByRole('button', { name: '会话详情' }).click()
  await expect(page.getByRole('heading', { name: '群公告' })).toBeVisible()
  await page.screenshot({ path: 'test-results/chat-details.png', fullPage: true, animations: 'disabled' })
  await page.getByRole('button', { name: '关闭会话详情' }).click()
  await page.getByRole('button', { name: '通讯录', exact: true }).click()
  await expect(page.getByRole('heading', { name: '我的好友' })).toBeVisible()
  await page.getByRole('button', { name: '个人资料', exact: true }).click()
  await expect(page.getByRole('dialog', { name: '关于我' })).toBeVisible()
  await page.getByLabel('昵称', { exact: true }).fill('予安')
  await page.getByRole('button', { name: '保存资料' }).click()
  await expect(page.getByRole('dialog')).toHaveCount(0)
  await page.getByRole('button', { name: '退出登录', exact: true }).click()
  await page.getByRole('button', { name: '退出预览', exact: true }).click()
  await expect(page.getByRole('heading', { name: '很高兴，又见面了' })).toBeVisible()
  expect(errors).toEqual([])
})

test('Mobile: list → conversation → back, no horizontal overflow', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/')
  await page.screenshot({ path: 'test-results/auth-mobile.png', fullPage: true, animations: 'disabled' })
  await page.getByRole('button', { name: '浏览界面预览' }).click()
  await expect(page.getByRole('textbox', { name: '消息内容' })).toBeVisible()
  await page.screenshot({ path: 'test-results/chat-mobile.png', fullPage: true, animations: 'disabled' })
  await page.getByRole('button', { name: '返回会话列表' }).click()
  await expect(page.getByRole('heading', { name: /^消息/ })).toBeVisible()
  await page.getByRole('button', { name: /林予安.*周末要不要/ }).click()
  await expect(page.getByRole('heading', { name: '林予安' })).toBeVisible()
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
})

async function fixture(page: Page, attachment?: { content: string; type: number }) {
  const frames: Record<string, unknown>[] = []
  let sent = false
  await page.route('**/api/**', route => {
    const path = new URL(route.request().url()).pathname
    const data = path.endsWith('/login') ? { code: 200, uid: '42', access_token: 'test-token', refresh_token: 'refresh' }
      : path.endsWith('/register') ? { code: 200, message: 'ok' }
      : path.endsWith('/heart') ? { code: 200, access_token: '' }
      : path.endsWith('/user/info') ? { name: '测试用户', photo: '', signature: 'hello' }
      : path.endsWith('/friend/list') ? { code: 200, list: [{ id: '43', friend_name: '测试好友' }] }
      : path.endsWith('/group/sessions') ? { code: 200, sessions: [{ group_id: '77', name: '集成群', last_seq: '1', last_read_seq: '0', visible_after_seq: '0', membership_version: '1', has_unread: true }] }
      : { code: 200, list: [] }
    return route.fulfill({ json: data })
  })
  await page.routeWebSocket('**/ws?*', ws => {
    ws.onMessage(raw => {
      const f = JSON.parse(String(raw)) as Record<string, unknown>; frames.push(f)
      const p = f.payload as Record<string, unknown>
      const respond = (type: string, data: unknown) => ws.send(JSON.stringify({ req_id: f.reqId, type, data }))
      const original = { msg_id: '1', group_id: '77', seq: '1', from_user_id: '43', client_msg_id: '99', msg_type: attachment?.type || 1, content: attachment?.content || '服务端已有消息', send_time: Date.now() }
      if (f.type === 'chat.GetNewGroupMsg') respond('chat.groupNewMsgBlock', { messages: p.cursor_seq === '0' ? [original] : [], next_cursor: sent ? '2' : '1', has_more: false, membership_version: '1', last_seq: sent ? '2' : '1' })
      if (f.type === 'chat.SendGroupMsg') {
        sent = true
        ws.send(JSON.stringify({ type: 'chat.groupMsg', data: { ...original, msg_id: '2', seq: '2', from_user_id: '42', client_msg_id: p.client_msg_id, content: p.content, notify: true } }))
        respond('ack.GroupMsg', { msg_id: '2', seq: '2', state: 'succeeded', client_msg_id: p.client_msg_id, send_time: Date.now() })
      }
      if (f.type === 'ack.GroupMsgRead') respond('ack.GroupRead', { seq: p.seq })
      if (f.type === 'chat.GetHistoryGroupMsg') respond('chat.groupHistoryMsgBlock', { messages: [], has_more: false, next_cursor: '0' })
      if (f.type === 'ai.CreateSession') respond('ai.Session', { session_id: '9223372036854775807' })
      if (f.type === 'ai.GetHistoryAiMsg') respond('ai.HistoryMsgBlock', [])
      if (f.type === 'ai.GetNewAiMsg') respond('ai.NewMsgBlock', [])
      if (f.type === 'ai.SendMsgToAi') {
        ws.send(JSON.stringify({ type: 'ai.LlmResponse.Ok', data: { task_id: 'task-1', session_id: p.session_id, client_msg_id: p.client_msg_id, msg_id: '9007199254740998', turn_seq: '1', state: 'succeeded', content: '模拟模型回复：计划已经整理好了。' } }))
        respond('ack.AiMsg', { session_id: p.session_id, task_id: 'task-1', client_msg_id: p.client_msg_id, msg_id: '9007199254740997', send_time: Date.now(), state: 'queued' })
      }
    })
  })
  return frames
}

test('Attachments: legacy path, login header, signed image and one automatic re-sign', async ({ page }) => {
  const key = 'ChatFile/picture/123.jpg'
  await fixture(page, { content: 'http://localhost:3001/objects/my-bucket/' + key, type: 2 })
  let signatures = 0
  const errors: string[] = []
  page.on('pageerror', e => errors.push(e.message))
  await page.route('**/api/file/download-url', route => {
    expect(route.request().method()).toBe('POST')
    expect(route.request().headers().authorization).toBe('Bearer test-token')
    expect(route.request().postDataJSON()).toEqual({ path: key })
    signatures++
    return route.fulfill({ json: { code: 200, path: key, url: `http://127.0.0.1:8333/my-bucket/${key}?X-Amz-Signature=test${signatures}`, expires_at: Math.floor(Date.now() / 1000) + 300 } })
  })
  await page.route('http://127.0.0.1:8333/**', route => route.request().url().endsWith('test1')
    ? route.fulfill({ status: 403, body: 'ExpiredToken' })
    : route.fulfill({ contentType: 'image/png', body: Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=', 'base64') }))
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await page.getByRole('button', { name: /集成群/ }).click()
  await expect(page.getByAltText('聊天图片')).toBeVisible()
  await expect.poll(() => page.getByAltText('聊天图片').evaluate(el => (el as HTMLImageElement).naturalWidth)).toBe(1)
  expect(signatures).toBe(2)
  expect(errors).toEqual([])
})

test('Attachments: regular files request a signed link only when clicked', async ({ page }) => {
  const key = 'ChatFile/file/123.pdf'
  await fixture(page, { content: key, type: 5 })
  let signatures = 0
  await page.route('**/api/file/download-url', route => { signatures++; return route.fulfill({ json: { code: 200, path: key, url: 'http://127.0.0.1:8333/my-bucket/' + key + '?X-Amz-Signature=test', expires_at: Math.floor(Date.now() / 1000) + 300 } }) })
  await page.context().route('http://127.0.0.1:8333/**', route => route.fulfill({ contentType: 'text/plain', body: 'attachment test' }))
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await page.getByRole('button', { name: /集成群/ }).click()
  await expect(page.getByRole('button', { name: /下载附件/ })).toBeVisible()
  expect(signatures).toBe(0)
  const popupPromise = page.waitForEvent('popup')
  await page.getByRole('button', { name: /下载附件/ }).click()
  const popup = await popupPromise
  await expect(popup).toHaveURL(/X-Amz-Signature=test/)
  expect(signatures).toBe(1)
  await popup.close()
})

test('Mock protocol: login, group sync, broadcast-before-ACK dedup and read watermark', async ({ page }) => {
  const frames = await fixture(page)
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await page.getByRole('button', { name: /集成群/ }).click()
  await expect(page.getByText('服务端已有消息', { exact: true }).last()).toBeVisible()
  await page.getByRole('textbox', { name: '消息内容' }).fill('协议检查')
  await page.getByRole('button', { name: '发送消息', exact: true }).click()
  await expect(page.locator('.message-bubble').filter({ hasText: '协议检查' })).toHaveCount(1)
  await expect(page.getByText('已发送', { exact: true })).toBeVisible()
  expect(frames.some(f => f.type === 'chat.GetNewGroupMsg')).toBe(true)
  expect(frames.some(f => f.type === 'ack.GroupMsgRead')).toBe(true)
  const outgoing = frames.find(f => f.type === 'chat.SendGroupMsg')?.payload as Record<string, unknown>
  expect(typeof outgoing.client_msg_id).toBe('string')
  expect(outgoing.from_user_id).toBeUndefined()
})

test('Registration validates confirmation and returns to login after success', async ({ page }) => {
  await fixture(page)
  await page.goto('/')
  await page.getByRole('button', { name: '加入栖语' }).click()
  await page.getByLabel('账号', { exact: true }).fill('new-user')
  await page.getByLabel('邮箱', { exact: true }).fill('demo@example.com')
  await page.getByLabel('密码', { exact: true }).fill('password123')
  await page.getByLabel('确认密码', { exact: true }).fill('notmatching')
  await page.getByRole('button', { name: '创建我的账号' }).click()
  await expect(page.getByRole('alert')).toContainText('两次输入的密码不一致')
  await page.getByLabel('确认密码', { exact: true }).fill('password123')
  await page.getByRole('button', { name: '创建我的账号' }).click()
  await expect(page.getByText('账号已创建，登录后开始新的交流。')).toBeVisible()
})

test('AI agent settings, avatar, new conversation binding and soft-delete UI', async ({ page }) => {
  const frames = await fixture(page)
  let saved = { agent_id: '0', agent_name: '', agent_prompt: '', agent_avatar: '' }, deleted = false
  await page.route('**/api/ai/**', route => {
    const path = new URL(route.request().url()).pathname
    expect(route.request().headers().authorization).toBe('Bearer test-token')
    if (path.endsWith('/agent/get')) return route.fulfill({ json: saved })
    if (path.endsWith('/agent/set')) { const body = route.request().postDataJSON() as typeof saved; saved = {...body, agent_id:'21'}; return route.fulfill({json:{code:200}}) }
    if (path.endsWith('/agent/avatar')) return route.fulfill({json:{code:200,path:'avatars/42/agent-test.png'}})
    if (path.endsWith('/session/delete')) { expect(route.request().postDataJSON()).toEqual({session_id:'9223372036854775807'}); deleted = true; return route.fulfill({json:{code:200}}) }
    return route.fulfill({json:{base:{code:200},list:[],page:{total:0}}})
  })
  await page.route('**/filer/avatars/42/agent-test.png',route=>route.fulfill({contentType:'image/png',body:Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=','base64')}))
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await page.getByRole('button', { name: 'AI 对话', exact:true }).click()
  await page.getByRole('button', { name: /智能体设置/ }).click()
  await page.getByLabel('角色名称').fill('灯塔书屋')
  await page.getByLabel('角色描述').fill('你是温和简练的守夜人阿澜。')
  await page.getByLabel('选择智能体头像').setInputFiles({name:'avatar.png',mimeType:'image/png',buffer:Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jRZkAAAAASUVORK5CYII=','base64')})
  await page.getByRole('button',{name:'保存智能体',exact:true}).click()
  await expect.poll(()=>saved.agent_id).toBe('21')
  expect(saved.agent_avatar).toBe('avatars/42/agent-test.png')
  await page.screenshot({path:'test-results/agent-settings-desktop.png',fullPage:true})
  await page.setViewportSize({width:390,height:844})
  expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true)
  await page.screenshot({path:'test-results/agent-settings-mobile.png',fullPage:true})
  await page.getByRole('button',{name:'返回 AI 列表'}).click()
  await page.getByRole('button',{name:'开启新对话',exact:true}).first().click()
  await expect.poll(()=>frames.some(f=>f.type==='ai.CreateSession' && (f.payload as Record<string,unknown>).agent_id==='21')).toBe(true)
  await page.getByRole('button',{name:'会话详情'}).click()
  await page.getByRole('button',{name:'删除 AI 对话',exact:true}).click()
  await page.getByRole('button',{name:'确认',exact:true}).click()
  await expect.poll(()=>deleted).toBe(true)
  await expect(page.getByRole('button',{name:/灯塔书屋.*从一句问候/})).toHaveCount(0)
})

test('Mock AI: async acknowledgement and result use exact session ID', async ({ page }) => {
  const frames = await fixture(page)
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await page.getByRole('button', { name: 'AI 对话', exact: true }).click()
  await page.getByRole('button', { name: '开启新对话', exact: true }).first().click()
  await expect(page.getByRole('heading', { name: /新对话/ })).toBeVisible()
  await page.getByRole('textbox', { name: '消息内容' }).fill('帮我整理计划')
  await page.getByRole('button', { name: '发送消息', exact: true }).click()
  await expect(page.locator('.message-bubble').filter({ hasText: '模拟模型回复：计划已经整理好了。' })).toHaveCount(1)
  const outgoing = frames.find(f => f.type === 'ai.SendMsgToAi')?.payload as Record<string, unknown>
  expect(outgoing.session_id).toBe('9223372036854775807')
  await expect(page.getByText('回复完成', { exact: true })).toBeVisible()
  await expect.poll(() => page.evaluate(() => localStorage.getItem('qiyu.cache.v1:42') || '')).toContain('9007199254740997')
})

test('Unavailable API surfaces an error without entering a fabricated account', async ({ page }) => {
  await page.route('**/api/user/login', route => route.fulfill({ status: 503, contentType: 'text/plain', body: '服务暂不可用' }))
  await page.goto('/')
  await page.getByLabel('账号', { exact: true }).fill('tester')
  await page.getByLabel('密码', { exact: true }).fill('test-password')
  await page.getByRole('button', { name: '登录，开启对话' }).click()
  await expect(page.getByRole('alert')).toContainText('服务暂不可用')
  await expect(page.getByRole('heading', { name: '很高兴，又见面了' })).toBeVisible()
})

test('AI streaming: immediate bubble, one-second sorted playback and final replacement', async ({page})=>{
  const errors:string[]=[];page.on('pageerror',e=>errors.push(e.message))
  await fixture(page)
  let push:(type:string,data:Record<string,unknown>)=>void=()=>{throw new Error('WS not ready')}
  let payload:Record<string,unknown>={}
  await page.routeWebSocket('**/ws?*',ws=>{
    push=(type,data)=>ws.send(JSON.stringify({type,data:{session_id:payload.session_id,client_msg_id:payload.client_msg_id,task_id:'stream-task',turn_seq:'1',attempt:1,...data}}))
    ws.onMessage(raw=>{
      const f=JSON.parse(String(raw)) as {type:string;reqId:string;payload:Record<string,unknown>}
      const respond=(type:string,data:unknown)=>ws.send(JSON.stringify({req_id:f.reqId,type,data}))
      if(f.type==='ai.CreateSession')respond('ai.Session',{session_id:'9223372036854775807'})
      if(f.type==='ai.GetHistoryAiMsg')respond('ai.HistoryMsgBlock',[])
      if(f.type==='ai.GetNewAiMsg')respond('ai.NewMsgBlock',[])
      if(f.type==='ai.SendMsgToAi'){payload=f.payload;respond('ack.AiMsg',{task_id:'stream-task',msg_id:'9007199254740997',state:'queued'})}
    })
  })
  await page.goto('/')
  await page.getByLabel('账号',{exact:true}).fill('tester');await page.getByLabel('密码',{exact:true}).fill('test-password')
  await page.getByRole('button',{name:'登录，开启对话'}).click()
  await page.getByRole('button',{name:'AI 对话',exact:true}).click()
  await page.getByRole('button',{name:'开启新对话',exact:true}).first().click()
  await page.clock.install()
  await page.getByRole('textbox',{name:'消息内容'}).fill('测试流式回答')
  await page.getByRole('button',{name:'发送消息',exact:true}).click()
  await expect(page.locator('.message-row:not(.own) .message-bubble')).toHaveCount(1)
  await expect(page.getByText('等待回复',{exact:true})).toBeVisible()
  await expect.poll(()=>payload.client_msg_id).toBeTruthy()
  push('ai.LlmResponse.Start',{seq:1,state:'running'})
  await expect(page.getByText('思考中',{exact:true})).toBeVisible()
  push('ai.LlmResponse.Delta',{seq:3,state:'running',content:'丙丁',reasoning_content:''})
  push('ai.LlmResponse.Delta',{seq:2,state:'running',content:'\n\n甲乙',reasoning_content:'先思考。'})
  // Let browser WS callbacks enqueue packets before advancing the virtual clock.
  await page.waitForTimeout(80)
  await page.clock.runFor(999)
  await expect(page.locator('.ai-reasoning')).toHaveCount(0)
  await page.clock.runFor(501)
  await expect(page.locator('.ai-reasoning')).toBeVisible()
  await page.clock.runFor(500)
  await expect(page.locator('.message-row:not(.own) .message-bubble')).toContainText('甲乙丙丁')
  expect(await page.locator('.message-row:not(.own) .message-bubble > p').textContent()).toBe('甲乙丙丁')
  await page.screenshot({path:'test-results/ai-streaming.png'})
  push('ai.LlmResponse.Delta',{seq:4,state:'running',content:'动画尚未播完的内容'})
  push('ai.LlmResponse.Ok',{state:'succeeded',msg_id:'9007199254740998',content:'\n\n最终权威答案',reasoning_content:'完整思考内容'})
  await expect(page.locator('.message-row:not(.own) .message-bubble')).toContainText('最终权威答案')
  expect(await page.locator('.message-row:not(.own) .message-bubble > p').textContent()).toBe('最终权威答案')
  await expect(page.locator('.ai-reasoning')).not.toHaveAttribute('open')
  await page.clock.runFor(2500)
  await expect(page.locator('.message-row:not(.own) .message-bubble')).not.toContainText('动画尚未播完')
  await expect(page.locator('.message-row:not(.own) .message-bubble')).toHaveCount(1)
  await page.locator('.ai-reasoning summary').click()
  await expect(page.getByText('完整思考内容',{exact:true})).toBeVisible()
  await page.screenshot({path:'test-results/ai-stream-final.png'})
  expect(errors).toEqual([])
})

import type { Conversation, Friend, Message } from './protocol'

export const demoFriends: Friend[] = [
  { id: '102', name: '林予安', remark: '', avatar: '', signature: '把日常过成喜欢的样子。', online: true },
  { id: '103', name: '陈一舟', remark: '', avatar: '', signature: '正在探索新的可能。', online: true },
  { id: '104', name: '许知夏', remark: '', avatar: '', signature: '今天也要去看看窗外。', online: false },
]
export function demoContent(): { conversations: Conversation[]; messages: Record<string, Message[]> } {
  const now = Date.now()
  const make = (id: string, name: string, kind: Conversation['kind'], preview: string, offset: number, unread = 0): Conversation => ({ key: `${kind}:${id}`, id, name, kind, preview, updated: now - offset, avatar: '', subtitle: kind === 'group' ? '共同创造，也分享日常' : '', online: true, unread, version: '1', visibleAfter: '0', readSeq: '5', lastSeq: '5' })
  const conversations = [
    make('301', '设计散步小组', 'group', '林予安：留一点空白，让想法自然发生。', 60000),
    make('102', '林予安', 'private', '周末要不要一起去看看新展？', 120000, 2),
    make('103', '陈一舟', 'private', '收到，我把资料整理好发你。', 1800000),
    make('302', '周末出逃计划', 'group', '许知夏：发现了一家很棒的咖啡店', 3600000, 1),
    make('401', '栖语 AI', 'ai', '一个可以慢慢展开想法的地方。', 7200000),
    make('104', '许知夏', 'private', '下次见，一路顺风。', 86400000),
  ]
  const row = (id: string, sender: string, content: string, offset: number): Message => ({ id, seq: id, sender, content, type: 1, time: now - offset })
  return { conversations, messages: {
    'group:301@1': [
      row('1', '102', '早上好！整理了一些关于「慢下来」的灵感，想和大家分享。', 1800000),
      row('2', '103', '正好，这个周末想找个地方散散步。\n不赶时间，只是看看街道、树影和路过的人。', 1600000),
      row('3', '101', '喜欢这个想法。好的交流，也需要一点从容。', 1400000),
      row('4', '104', '那就把周六下午留给彼此吧。带上相机，也带上好心情。', 800000),
      row('5', '102', '留一点空白，让想法自然发生。', 60000),
    ],
    'private:102': [row('10', '102', '周末要不要一起去看看新展？', 120000), row('11', '102', '听说有一个关于城市与自然的摄影展，应该会喜欢。', 110000)],
    'private:103': [row('12', '103', '收到，我把资料整理好发你。', 1800000)],
    'group:302@1': [row('1', '104', '发现了一家很棒的咖啡店，下次一起去！', 3600000)],
    'ai:401': [row('20', 'ai', '你好，我是栖语 AI。\n你可以和我讨论一个想法、整理一段文字，或只是聊聊今天。', 7200000)],
    'private:104': [row('21', '104', '下次见，一路顺风。', 86400000)],
  } }
}

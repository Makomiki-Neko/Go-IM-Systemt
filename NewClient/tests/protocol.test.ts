import test from 'node:test'
import assert from 'node:assert/strict'
import { decode, object, messageFrom, mergeMessages, idCompare } from '../src/lib/protocol'

test('AI numeric IDs never round through JSON parsing', () => {
  const r = object(decode('{"MsgID":9007199254740997,"SessionID":9223372036854775807,"MsgType":1,"TurnSeq":2}'))
  assert.equal(r.MsgID, '9007199254740997')
  assert.equal(r.SessionID, '9223372036854775807')
  assert.equal(r.MsgType, 1)
  assert.equal(messageFrom(r, 'ai').id, '9007199254740997')
})
test('Broadcast before ACK replaces optimistic message by immutable client ID', () => {
  const local = { id: 'local-123', clientId: '123', sender: '42', type: 1, time: 1, content: 'hello', state: 'sending' as const }
  const server = { ...local, id: '88', seq: '5', time: 2, state: 'succeeded' as const }
  const result = mergeMessages([local], [server, server])
  assert.equal(result.length, 1)
  assert.equal(result[0].id, '88')
  assert.equal(result[0].state, 'succeeded')
})
test('Same client ID from different group members is not deduplicated', () => {
  const m = { id: '1', clientId: '123', sender: '42', type: 1, time: 1, content: 'a', seq: '9007199254740998' }
  const result = mergeMessages([m], [{ ...m, id: '2', sender: '43', seq: '9007199254740997' }])
  assert.equal(result.length, 2)
  assert.equal(result[0].id, '2')
  assert.equal(idCompare('9007199254740998', '9007199254740997'), 1)
})
test('Private seconds and group milliseconds normalize to same time', () => {
  assert.equal(messageFrom({ send_time: 1790000000 }, 'private').time, messageFrom({ send_time: 1790000000000 }, 'group').time)
})
test('AI user history does not change a failed task into a success', () => {
  const failed = { id: '10', sender: '42', content: 'test', type: 1, time: 1, taskId: 'task', state: 'failed' as const }
  const row = messageFrom({ MsgID: 10, UserID: 42, Content: 'test', MsgType: 1 }, 'ai')
  assert.equal(mergeMessages([failed], [row])[0].state, 'failed')
})

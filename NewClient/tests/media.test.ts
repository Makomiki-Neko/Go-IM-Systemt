import test from 'node:test'
import assert from 'node:assert/strict'
import { chatFilePath } from '../src/lib/media'
test('Stored object keys and legacy URLs normalize to paths without signatures or hosts', () => {
  const key = 'ChatFile/picture/a31b03b1-8df7-45f3-8201-1e827da6b380.jpg'
  for (const input of [key, '/' + key, 'http://localhost:3001/objects/my-bucket/' + key, 'http://localhost:8333/my-bucket/' + key + '?X-Amz-Signature=old', '/filer/buckets/my-bucket/' + key]) assert.equal(chatFilePath(input), key)
})
test('Reject arbitrary resources, other buckets, directory paths and encoded traversal', () => {
  for (const input of ['http://evil.invalid/private', 'javascript:alert(1)', '/objects/other/ChatFile/picture/a.jpg', 'ChatFile/picture/../x', 'ChatFile/picture/a%2fb', 'ChatFile/picture/', '/filer/buckets/my-bucket/']) assert.equal(chatFilePath(input), '')
})

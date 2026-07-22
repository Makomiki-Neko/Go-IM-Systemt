const AVATAR_BASE = 'http://localhost:8890/buckets/my-bucket'

export function getAvatarUrl(photo?: string): string {
  if (!photo) return ''
  if (photo.startsWith('http://') || photo.startsWith('https://')) return photo
  return `${AVATAR_BASE}/${photo.replace(/^\//, '')}`
}

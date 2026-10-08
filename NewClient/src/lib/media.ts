// Normalize old stored URLs locally. Never send credentials to a URL from message content.
export function chatFilePath(value: string): string {
  let key = value
  if (/^https?:\/\//i.test(key) || key.startsWith('/')) {
    try {
      key = decodeURIComponent(new URL(key, 'http://local.invalid').pathname)
      key = key.replace(/^\/(?:objects\/|filer\/buckets\/|buckets\/)?my-bucket\//, '').replace(/^\/ChatFile\//, 'ChatFile/')
    } catch { return '' }
  }
  return /^ChatFile\/(picture|voice|audio|video|file)\/[A-Za-z0-9_-][A-Za-z0-9._-]{0,255}$/.test(key) ? key : ''
}

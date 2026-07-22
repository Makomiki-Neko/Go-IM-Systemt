import http from './http'
import type { GetUnreadNumberResp } from '@/types'

export const getUnreadPrivateMsgNumberApi = () =>
  http.get<GetUnreadNumberResp>('/chat/GetUnreadPrivateMsgNumber')

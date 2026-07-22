import http from './http'
import type {
  SearchUserResp, SendFriendReq, SendFriendResp,
  HandleFriendReq, DeleteFriendReq, SetRemarkReq,
  GetFriendListResp, CommonResp,
} from '@/types'

/** 将对象转为 URLSearchParams（用于 form 标签类 API） */
function toForm(data: Record<string, any>): URLSearchParams {
  const params = new URLSearchParams()
  for (const [k, v] of Object.entries(data)) {
    if (v !== undefined && v !== null) {
      params.append(k, String(v))
    }
  }
  return params
}

const formHeaders = { 'Content-Type': 'application/x-www-form-urlencoded' }

/* ---- 以下 API 服务端使用 form 标签 ---- */

export const searchUserApi = (searchName: string, page = 1, size = 20) =>
  http.post<SearchUserResp>('/friend/search', toForm({ search_name: searchName, page, size }), {
    headers: formHeaders,
  })

export const getFriendListApi = (page = 1, size = 200) =>
  http.post<GetFriendListResp>('/friend/list', toForm({ page, size }), {
    headers: formHeaders,
  })

export const getFriendApplyListApi = (page = 1, size = 200) =>
  http.post<GetFriendListResp>('/friend/apply_list', toForm({ page, size }), {
    headers: formHeaders,
  })

/* ---- 以下 API 服务端使用 json 标签，直接发送 JSON ---- */

export const sendFriendRequestApi = (data: SendFriendReq) =>
  http.post<SendFriendResp>('/friend/request', data)

export const handleFriendRequestApi = (data: HandleFriendReq) =>
  http.post<CommonResp>('/friend/request_handle', data)

export const deleteFriendApi = (data: DeleteFriendReq) =>
  http.post<CommonResp>('/friend/delete', data)

export const setFriendRemarkApi = (data: SetRemarkReq) =>
  http.post<CommonResp>('/friend/remark', data)

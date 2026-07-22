/* ---- REST API 类型 (与 Backend/api/api.api 对齐) ---- */
export interface LoginRequest {
  account: string
  password: string
  platform: string
  device_id: string
}
export interface LoginResponse {
  code: number; uid: string; access_token: string; refresh_token: string; message: string
}
export interface RegisterRequest {
  account: string; password: string; email: string
}
export interface RegisterResponse {
  code: number; message: string
}
export interface HeartRequest {
  uid: string; account: string; refresh_token: string; device_id: string; platform: string
}
export interface HeartResponse {
  code: number; access_token: string
}
export interface UserInfoResponse {
  name: string; photo: string; gender: string; signature: string
  birthday: number; createdtime: number; lastonline: number; online: boolean
}
export interface UpdateUserInfoRequest {
  name: string; photo: string; gender: string; birthday: number; signature: string
}
export interface UpdateAvatarResponse {
  code: number; photo_id: string
}

/* ---- Friend 类型 ---- */
export interface FriendInfo {
  id: string; friend_name: string; remark: string; avatar: string
  signature: string; gender: string; birthday: number; last_online: number
  online_statu: boolean; status: number; created_at: number
}
export interface CommonResp {
  code: number; msg: string
}
export interface SearchUserResp extends CommonResp {
  list: FriendInfo[]; total: number; page: number; size: number
}
export interface SendFriendReq {
  friend_id: string; msg: string
}
export interface SendFriendResp extends CommonResp {
  request_id: string
}
export interface HandleFriendReq {
  request_id: string; accept: boolean; sender_id: string; user_id: string
}
export interface DeleteFriendReq {
  friend_id: string
}
export interface SetRemarkReq {
  friend_id: string; remark: string
}
export interface GetFriendListResp extends CommonResp {
  list: FriendInfo[]; total: number; page: number; size: number
}

/* ---- Chat 类型 ---- */
export interface UnReadInfo {
  id: string; count: number
}
export interface GetUnreadNumberResp {
  list: UnReadInfo[]
}

/* ---- WS 消息类型 (与 Backend/gateway 及 Project.MD 对齐) ---- */
export interface WsMessage {
  type: string
  reqId: number
  payload?: any
}

export interface WsPushMessage {
  req_id: number
  type: string
  data: any
}

export interface ChatMessage {
  msg_id: string
  from_user_id: string
  to_user_id: string
  msg_type: number    // 1文本 2图片 3语音 4视频 5文件
  content: string
  send_time: number
  status: number      // 0已发送 1已送达 2已读
  _sending?: boolean   // 前端乐观标识：正在发送
  _failed?: boolean    // 前端乐观标识：发送失败
}

export interface Conversation {
  userId: string
  friendName: string
  remark: string
  avatar: string
  lastMsg: string
  lastTime: number
  unread: number
  isOnline: boolean
  lastMsgId: string      // 会话最后收到的消息ID，用于拉取新消息和历史
  pulledThisSession: boolean  // 本次登录/重连后是否已拉取过新消息
}

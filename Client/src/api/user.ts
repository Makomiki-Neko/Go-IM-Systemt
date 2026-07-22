import http from './http'
import type { UserInfoResponse, UpdateUserInfoRequest, UpdateAvatarResponse } from '@/types'

export const getUserInfoApi = (account: string) =>
  http.get<UserInfoResponse>('/user/info', { params: { account } })

export const updateUserInfoApi = (data: UpdateUserInfoRequest) =>
  http.post<{ code: number }>('/user/update/info', data)

export const updateAvatarApi = (formData: FormData) =>
  http.post<UpdateAvatarResponse>('/user/update/avatar', formData)

export const updateEmailApi = (account: string, email: string) =>
  http.post<{ code: number }>('/user/update/email', { account, email })

export const updatePasswordApi = (account: string, password: string) =>
  http.post<{ code: number }>('/user/update/password', { account, password })

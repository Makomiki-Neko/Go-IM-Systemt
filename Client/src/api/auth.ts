import axios from 'axios'
import http from './http'
import type { LoginRequest, LoginResponse, RegisterRequest, RegisterResponse, HeartRequest, HeartResponse } from '@/types'

export const loginApi = (data: LoginRequest) =>
  http.post<LoginResponse>('/user/login', data)

export const registerApi = (data: RegisterRequest) =>
  http.post<RegisterResponse>('/user/register', data)

export const heartApi = (data: HeartRequest) => {
  const token = localStorage.getItem('access_token')
  return axios.post<HeartResponse>('/api/user/heart', data, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
  })
}

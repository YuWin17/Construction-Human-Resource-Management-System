import { apiClient, tokenStorageKey } from './client'

export interface AdminProfile {
  id: string
  username: string
}

interface ApiEnvelope<T> {
  data: T
}

interface LoginPayload {
  access_token: string
  admin: AdminProfile
}

// 登录管理员，并在当前浏览器会话中保存短期访问令牌。
export async function login(username: string, password: string): Promise<AdminProfile> {
  const response = await apiClient.post<ApiEnvelope<LoginPayload>>('/auth/login', {
    username,
    password,
  })
  sessionStorage.setItem(tokenStorageKey, response.data.data.access_token)
  return response.data.data.admin
}

// 确认本地访问令牌仍有效，并返回其所属管理员。
export async function getCurrentAdmin(): Promise<AdminProfile> {
  const response = await apiClient.get<ApiEnvelope<AdminProfile>>('/auth/me')
  return response.data.data
}

// 即使服务端退出请求失败，也始终清除本地认证状态。
export async function logout(): Promise<void> {
  try {
    await apiClient.post('/auth/logout')
  } finally {
    sessionStorage.removeItem(tokenStorageKey)
  }
}

export function hasAccessToken(): boolean {
  return Boolean(sessionStorage.getItem(tokenStorageKey))
}

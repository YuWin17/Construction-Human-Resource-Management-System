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

export async function login(username: string, password: string): Promise<AdminProfile> {
  const response = await apiClient.post<ApiEnvelope<LoginPayload>>('/auth/login', {
    username,
    password,
  })
  sessionStorage.setItem(tokenStorageKey, response.data.data.access_token)
  return response.data.data.admin
}

export async function getCurrentAdmin(): Promise<AdminProfile> {
  const response = await apiClient.get<ApiEnvelope<AdminProfile>>('/auth/me')
  return response.data.data
}

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

import axios from 'axios'

export const tokenStorageKey = 'construction-hrms.access-token'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1',
  timeout: 15_000,
})

apiClient.interceptors.request.use((request) => {
  const token = sessionStorage.getItem(tokenStorageKey)
  if (token) {
    request.headers.Authorization = `Bearer ${token}`
  }
  return request
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      sessionStorage.removeItem(tokenStorageKey)
      if (window.location.pathname !== '/login') {
        window.location.assign('/login')
      }
    }
    return Promise.reject(error)
  },
)

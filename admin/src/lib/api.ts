import axios from 'axios'
import { useAuth } from '../composables/useAuth'

export type FeedbackStatus = 'open' | 'resolved'

export interface Feedback {
  id: string
  title: string
  content: string
  contact: string
  project: string
  pageUrl: string
  userAgent: string
  ip: string
  status: FeedbackStatus
  createdAt: string
  updatedAt: string
  resolvedAt?: string | null
}

export interface FeedbackCounts {
  total: number
  open: number
  resolved: number
}

export interface FeedbackList {
  items: Feedback[]
  total: number
  page: number
  pageSize: number
  counts: FeedbackCounts
}

export interface FeedbackQuery {
  page?: number
  pageSize?: number
  q?: string
  status?: string
  project?: string
}

export interface LoginResponse {
  token: string
  username: string
  mustChangePassword: boolean
}

export interface AdminInfo {
  username: string
  mustChangePassword: boolean
}

const client = axios.create({
  baseURL: '/api',
  timeout: 20000,
  headers: { Accept: 'application/json' },
})

client.interceptors.request.use((config) => {
  const { state } = useAuth()
  if (state.token) {
    config.headers.Authorization = `Bearer ${state.token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const requestUrl = String(error.config?.url ?? '')
    const isLoginRequest = requestUrl.includes('/admin/login')

    if (status === 401 && !isLoginRequest) {
      useAuth().clearSession()
      if (window.location.pathname !== '/login') {
        window.location.assign('/login')
      }
    }

    if (status === 403 && error.response?.data?.code === 'password_change_required') {
      useAuth().markMustChange()
    }

    return Promise.reject(error)
  },
)

export function errorMessage(error: unknown, fallback = '操作失败，请稍后重试'): string {
  if (axios.isAxiosError(error)) {
    const message = error.response?.data?.message
    if (typeof message === 'string' && message.trim()) {
      return message
    }
    if (error.code === 'ECONNABORTED') {
      return '请求超时，请检查网络连接'
    }
    if (!error.response) {
      return '无法连接服务，请确认后端已启动'
    }
  }
  return fallback
}

export function isCanceledError(error: unknown): boolean {
  return axios.isCancel(error) || (error as { name?: string })?.name === 'CanceledError'
}

export function login(payload: { username: string; password: string }): Promise<LoginResponse> {
  return client.post<LoginResponse>('/admin/login', payload).then((response) => response.data)
}

export function fetchMe(): Promise<AdminInfo> {
  return client.get<AdminInfo>('/admin/me').then((response) => response.data)
}

export function changePassword(payload: { oldPassword: string; newPassword: string }): Promise<LoginResponse> {
  return client.post<LoginResponse>('/admin/password', payload).then((response) => response.data)
}

export function logout(): Promise<void> {
  return client.post('/admin/logout').then(() => undefined)
}

export function fetchFeedbackList(query: FeedbackQuery, signal?: AbortSignal): Promise<FeedbackList> {
  const params: Record<string, string | number> = {
    page: query.page ?? 1,
    pageSize: query.pageSize ?? 20,
  }
  if (query.q) params.q = query.q
  if (query.status) params.status = query.status
  if (query.project) params.project = query.project

  return client.get<FeedbackList>('/admin/feedback', { params, signal }).then((response) => response.data)
}

export function fetchFeedbackDetail(id: string): Promise<Feedback> {
  return client.get<Feedback>(`/admin/feedback/${id}`).then((response) => response.data)
}

export function updateFeedbackStatus(id: string, status: FeedbackStatus): Promise<Feedback> {
  return client.patch<Feedback>(`/admin/feedback/${id}`, { status }).then((response) => response.data)
}

export function deleteFeedback(id: string): Promise<void> {
  return client.delete(`/admin/feedback/${id}`).then(() => undefined)
}

export function fetchProjects(): Promise<string[]> {
  return client.get<{ projects: string[] }>('/admin/projects').then((response) => response.data.projects)
}

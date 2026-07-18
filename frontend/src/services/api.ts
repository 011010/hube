import axios from 'axios'
import type { Task, CalendarEvent, App, WishlistItem, Diagram } from '../types'

export const API_BASE = (import.meta.env.VITE_API_URL ?? '/api/v1').replace(/\/$/, '')
export const TOKEN_KEY = 'hube_token'
export const http = axios.create({ baseURL: API_BASE })

// Attach the stored bearer token to every outgoing request.
http.interceptors.request.use(config => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// A 401 means the stored token is missing or no longer valid: drop it and
// bounce back to the auth gate with a single page reload. The module-level
// flag (plus the hadToken check) guarantees we never reload in a loop.
let reloadingForUnauthorized = false

http.interceptors.response.use(
  response => response,
  (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const hadToken = localStorage.getItem(TOKEN_KEY) !== null
      localStorage.removeItem(TOKEN_KEY)
      if (hadToken && !reloadingForUnauthorized) {
        reloadingForUnauthorized = true
        window.location.reload()
      }
    }
    return Promise.reject(error)
  },
)

export const tasksApi = {
  list: () => http.get<Task[]>('/tasks').then(r => r.data),
  get: (id: string) => http.get<Task>(`/tasks/${id}`).then(r => r.data),
  create: (data: Omit<Task, 'id' | 'created_at' | 'updated_at'>) =>
    http.post<Task>('/tasks', data).then(r => r.data),
  update: (id: string, data: Partial<Task>) =>
    http.put<Task>(`/tasks/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/tasks/${id}`),
}

export const eventsApi = {
  list: (from?: string, to?: string) => {
    const params = from && to ? { from, to } : {}
    return http.get<CalendarEvent[]>('/events', { params }).then(r => r.data)
  },
  get: (id: string) => http.get<CalendarEvent>(`/events/${id}`).then(r => r.data),
  create: (data: Omit<CalendarEvent, 'id' | 'created_at' | 'updated_at'>) =>
    http.post<CalendarEvent>('/events', data).then(r => r.data),
  update: (id: string, data: Partial<CalendarEvent>) =>
    http.put<CalendarEvent>(`/events/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/events/${id}`),
}

export const appsApi = {
  list: () => http.get<App[]>('/apps').then(r => r.data),
  create: (data: Omit<App, 'id'>) => http.post<App>('/apps', data).then(r => r.data),
  update: (id: string, data: Partial<App>) =>
    http.put<App>(`/apps/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/apps/${id}`),
}

export const diagramsApi = {
  list: () => http.get<Diagram[]>('/diagrams').then(r => r.data),
  get: (id: string) => http.get<Diagram>(`/diagrams/${id}`).then(r => r.data),
  create: (data: Pick<Diagram, 'name'>) =>
    http.post<Diagram>('/diagrams', data).then(r => r.data),
  update: (id: string, data: Partial<Diagram>) =>
    http.put<Diagram>(`/diagrams/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/diagrams/${id}`),
}

export const wishlistApi = {
  list: () => http.get<WishlistItem[]>('/wishlist').then(r => r.data),
  get: (id: string) => http.get<WishlistItem>(`/wishlist/${id}`).then(r => r.data),
  create: (data: Omit<WishlistItem, 'id' | 'created_at' | 'updated_at'>) =>
    http.post<WishlistItem>('/wishlist', data).then(r => r.data),
  update: (id: string, data: Partial<WishlistItem>) =>
    http.put<WishlistItem>(`/wishlist/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/wishlist/${id}`),
}

import axios from 'axios'
import type { Task, CalendarEvent, App, WishlistItem } from '../types'

const http = axios.create({ baseURL: import.meta.env.VITE_API_URL ?? '/api/v1' })

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

export const wishlistApi = {
  list: () => http.get<WishlistItem[]>('/wishlist').then(r => r.data),
  get: (id: string) => http.get<WishlistItem>(`/wishlist/${id}`).then(r => r.data),
  create: (data: Omit<WishlistItem, 'id' | 'created_at' | 'updated_at'>) =>
    http.post<WishlistItem>('/wishlist', data).then(r => r.data),
  update: (id: string, data: Partial<WishlistItem>) =>
    http.put<WishlistItem>(`/wishlist/${id}`, data).then(r => r.data),
  delete: (id: string) => http.delete(`/wishlist/${id}`),
}

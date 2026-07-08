import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import type { Note, Folder } from '../types'

const api = axios.create({ baseURL: '/api/v1' })

export type NoteInput = Partial<Omit<Note, 'id' | 'created_at' | 'updated_at'>>
export type NoteUpdate = Partial<Note>

export function useNotes(folderID?: string) {
  return useQuery<Note[]>({
    queryKey: ['notes', folderID ?? 'all'],
    queryFn: () => api.get('/notes', { params: folderID ? { folder_id: folderID } : {} }).then(r => r.data),
  })
}

export function useNote(id: string | null) {
  return useQuery<Note>({
    queryKey: ['note', id],
    queryFn: () => api.get(`/notes/${id}`).then(r => r.data),
    enabled: Boolean(id),
  })
}

export function useSearchNotes(query: string) {
  return useQuery<Note[]>({
    queryKey: ['notes', 'search', query],
    queryFn: () => api.get('/notes/search', { params: { q: query } }).then(r => r.data),
    enabled: query.length > 1,
  })
}

export interface SemanticResult {
  note: Note
  score: number
}

export function useSemanticSearch(query: string) {
  return useQuery<SemanticResult[]>({
    queryKey: ['notes', 'semantic', query],
    queryFn: () =>
      api.post<SemanticResult[]>('/notes/semantic-search', { q: query, top_k: 5 }).then(r => r.data),
    enabled: query.length > 2,
    retry: false,
  })
}

export function useCreateNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: NoteInput) => api.post<Note>('/notes', data).then(r => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notes'] }),
  })
}

export function useUpdateNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: NoteUpdate }) =>
      api.put<Note>(`/notes/${id}`, data).then(r => r.data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['notes'] })
      qc.invalidateQueries({ queryKey: ['note', id] })
    },
  })
}

export function useDeleteNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/notes/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notes'] }),
  })
}

export function useFolders() {
  return useQuery<Folder[]>({
    queryKey: ['folders'],
    queryFn: () => api.get('/folders').then(r => r.data),
  })
}

export function useCreateFolder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; parent_id?: string }) => api.post<Folder>('/folders', data).then(r => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['folders'] }),
  })
}

export function useDeleteFolder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/folders/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['folders'] }),
  })
}

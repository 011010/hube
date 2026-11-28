import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { http } from '../services/api'
import { createCrudHooks } from './createCrudHooks'
import type { Note, Folder, NoteProperty, ClipInput } from '../types'

export type NoteInput = Partial<Omit<Note, 'id' | 'created_at' | 'updated_at' | 'properties'>>
export type NoteUpdate = Partial<Omit<Note, 'properties'>>

const notesCrud = createCrudHooks<Note, NoteInput, NoteUpdate>(
  'notes',
  {
    list: () => http.get<Note[]>('/notes').then(r => r.data),
    create: (data: NoteInput) => http.post<Note>('/notes', data).then(r => r.data),
    update: (id: string, data: NoteUpdate) => http.put<Note>(`/notes/${id}`, data).then(r => r.data),
    delete: (id: string) => http.delete(`/notes/${id}`),
  },
  { itemQueryKey: id => ['note', id] },
)

export function useNotes(folderID?: string, enabled = true) {
  return useQuery<Note[]>({
    queryKey: ['notes', folderID ?? 'all'],
    queryFn: () => http.get('/notes', { params: folderID ? { folder_id: folderID } : {} }).then(r => r.data),
    enabled,
  })
}

export function useNote(id: string | null) {
  return useQuery<Note>({
    queryKey: ['note', id],
    queryFn: () => http.get(`/notes/${id}`).then(r => r.data),
    enabled: Boolean(id),
  })
}

export function useSearchNotes(query: string) {
  return useQuery<Note[]>({
    queryKey: ['notes', 'search', query],
    queryFn: () => http.get('/notes/search', { params: { q: query } }).then(r => r.data),
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
      http.post<SemanticResult[]>('/notes/semantic-search', { q: query, top_k: 5 }).then(r => r.data),
    enabled: query.length > 2,
    retry: false,
  })
}

export const useCreateNote = notesCrud.useCreate
export const useUpdateNote = notesCrud.useUpdate
export const useDeleteNote = notesCrud.useDelete

export function useSetNoteProperties() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, properties }: { id: string; properties: NoteProperty[] }) =>
      http.put<NoteProperty[]>(`/notes/${id}/properties`, { properties }).then(r => r.data),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['note', variables.id] })
      qc.invalidateQueries({ queryKey: ['notes'] })
    },
  })
}

// Clips a web page into a new note. A 409 means the URL was clipped before
// and the error body carries `existing_id`; the caller decides what to do
// with it. Invalidations cover both the list and the single-note queries so
// a just-clipped note renders without waiting for a refetch.
export function useClipNote() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (input: ClipInput) => http.post<Note>('/notes/clip', input).then(r => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notes'] })
      qc.invalidateQueries({ queryKey: ['note'] })
    },
  })
}

export type GraphNodeType = 'note' | 'task' | 'project'
export type GraphEdgeType = 'link' | 'task' | 'project'

export interface NoteGraphNode {
  id: string
  label: string
  type: GraphNodeType
}

export interface NoteGraphEdge {
  source: string
  target: string
  type: GraphEdgeType
}

export interface NoteGraph {
  nodes: NoteGraphNode[]
  edges: NoteGraphEdge[]
}

export function useNoteGraph() {
  return useQuery<NoteGraph>({
    queryKey: ['notes', 'graph'],
    queryFn: () => http.get('/notes/graph').then(r => r.data),
  })
}

export function useFolders() {
  return useQuery<Folder[]>({
    queryKey: ['folders'],
    queryFn: () => http.get('/folders').then(r => r.data),
  })
}

export function useCreateFolder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { name: string; parent_id?: string }) => http.post<Folder>('/folders', data).then(r => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['folders'] }),
  })
}

export function useDeleteFolder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => http.delete(`/folders/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['folders'] }),
  })
}

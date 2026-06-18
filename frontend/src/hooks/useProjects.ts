import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'
import type { Project, Task } from '../types'

const api = axios.create({ baseURL: '/api/v1' })

export function useProjects() {
  return useQuery<Project[]>({
    queryKey: ['projects'],
    queryFn: () => api.get('/projects').then(r => r.data),
  })
}

export function useProject(id: string | null) {
  return useQuery<Project>({
    queryKey: ['project', id],
    queryFn: () => api.get(`/projects/${id}`).then(r => r.data),
    enabled: Boolean(id),
  })
}

export function useCreateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: Partial<Project>) => api.post<Project>('/projects', data).then(r => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  })
}

export function useUpdateProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Project> }) =>
      api.put<Project>(`/projects/${id}`, data).then(r => r.data),
    onSuccess: (_, { id }) => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.invalidateQueries({ queryKey: ['project', id] })
    },
  })
}

export function useDeleteProject() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: string) => api.delete(`/projects/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  })
}

export function useProjectTasks(projectId: string) {
  return useQuery<Task[]>({
    queryKey: ['tasks', 'project', projectId],
    queryFn: () => api.get('/tasks', { params: { project_id: projectId } }).then(r => r.data),
    enabled: Boolean(projectId),
  })
}

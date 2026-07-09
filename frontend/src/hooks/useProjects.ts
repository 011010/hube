import { useQuery } from '@tanstack/react-query'
import { http } from '../services/api'
import { createCrudHooks } from './createCrudHooks'
import type { Project, Task } from '../types'

const projectsCrud = createCrudHooks<Project, Partial<Project>, Partial<Project>>(
  'projects',
  {
    list: () => http.get<Project[]>('/projects').then(r => r.data),
    create: (data: Partial<Project>) => http.post<Project>('/projects', data).then(r => r.data),
    update: (id: string, data: Partial<Project>) => http.put<Project>(`/projects/${id}`, data).then(r => r.data),
    delete: (id: string) => http.delete(`/projects/${id}`),
  },
  { itemQueryKey: id => ['project', id] },
)

export const useProjects = projectsCrud.useList
export const useCreateProject = projectsCrud.useCreate
export const useUpdateProject = projectsCrud.useUpdate
export const useDeleteProject = projectsCrud.useDelete

export function useProject(id: string | null) {
  return useQuery<Project>({
    queryKey: ['project', id],
    queryFn: () => http.get(`/projects/${id}`).then(r => r.data),
    enabled: Boolean(id),
  })
}

export function useProjectTasks(projectId: string) {
  return useQuery<Task[]>({
    queryKey: ['tasks', 'project', projectId],
    queryFn: () => http.get('/tasks', { params: { project_id: projectId } }).then(r => r.data),
    enabled: Boolean(projectId),
  })
}

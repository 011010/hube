import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tasksApi } from '../services/api'
import type { Task } from '../types'

export function useTasks() {
  return useQuery({ queryKey: ['tasks'], queryFn: tasksApi.list })
}

export function useCreateTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: tasksApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useUpdateTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Task> }) =>
      tasksApi.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

export function useDeleteTask() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: tasksApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks'] }),
  })
}

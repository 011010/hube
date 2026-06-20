import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { diagramsApi } from '../services/api'
import type { Diagram } from '../types'

export function useDiagrams() {
  return useQuery({ queryKey: ['diagrams'], queryFn: diagramsApi.list })
}

export function useCreateDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: diagramsApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['diagrams'] }),
  })
}

export function useUpdateDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<Diagram> }) =>
      diagramsApi.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['diagrams'] }),
  })
}

export function useDeleteDiagram() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: diagramsApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['diagrams'] }),
  })
}

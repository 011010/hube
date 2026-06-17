import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { eventsApi } from '../services/api'
import type { CalendarEvent } from '../types'

export function useEvents(from: string, to: string) {
  return useQuery({
    queryKey: ['events', from, to],
    queryFn: () => eventsApi.list(from, to),
    enabled: !!from && !!to,
  })
}

export function useCreateEvent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: eventsApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  })
}

export function useUpdateEvent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<CalendarEvent> }) =>
      eventsApi.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  })
}

export function useDeleteEvent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: eventsApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  })
}

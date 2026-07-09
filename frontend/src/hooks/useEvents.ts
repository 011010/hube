import { useQuery } from '@tanstack/react-query'
import { createCrudHooks } from './createCrudHooks'
import { eventsApi } from '../services/api'
import type { CalendarEvent } from '../types'

const eventsCrud = createCrudHooks<
  CalendarEvent,
  Omit<CalendarEvent, 'id' | 'created_at' | 'updated_at'>,
  Partial<CalendarEvent>
>('events', eventsApi)

export function useEvents(from: string, to: string) {
  return useQuery({
    queryKey: ['events', from, to],
    queryFn: () => eventsApi.list(from, to),
    enabled: !!from && !!to,
  })
}

export const useCreateEvent = eventsCrud.useCreate
export const useUpdateEvent = eventsCrud.useUpdate
export const useDeleteEvent = eventsCrud.useDelete

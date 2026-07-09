import { useQuery } from '@tanstack/react-query'
import { http } from '../services/api'
import type { CardTrackerSummary } from '../types'

export function useCardTrackerSummary() {
  return useQuery<CardTrackerSummary>({
    queryKey: ['cards', 'summary'],
    queryFn: () => http.get('/cards/summary').then(r => r.data),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

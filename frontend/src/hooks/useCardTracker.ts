import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import type { CardTrackerSummary } from '../types'

export function useCardTrackerSummary() {
  return useQuery<CardTrackerSummary>({
    queryKey: ['cards', 'summary'],
    queryFn: () => axios.get('/api/v1/cards/summary').then(r => r.data),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

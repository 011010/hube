import { useQuery } from '@tanstack/react-query'
import axios from 'axios'
import type { FinanceSummary } from '../types'

export function useFinanceSummary() {
  return useQuery<FinanceSummary>({
    queryKey: ['finance', 'summary'],
    queryFn: () => axios.get('/api/v1/finance/summary').then(r => r.data),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

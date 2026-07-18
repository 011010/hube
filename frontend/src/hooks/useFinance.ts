import { useQuery } from '@tanstack/react-query'
import { http } from '../services/api'
import type { FinanceSummary } from '../types'

export function useFinanceSummary() {
  return useQuery<FinanceSummary>({
    queryKey: ['finance', 'summary'],
    queryFn: () => http.get('/finance/summary').then(r => r.data),
    staleTime: 5 * 60 * 1000,
    retry: false,
  })
}

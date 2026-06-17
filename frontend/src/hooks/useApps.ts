import { useQuery } from '@tanstack/react-query'
import { appsApi } from '../services/api'

export function useApps() {
  return useQuery({ queryKey: ['apps'], queryFn: appsApi.list })
}

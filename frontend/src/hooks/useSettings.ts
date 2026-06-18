import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })

export interface GeneralSettings {
  display_name: string
}

export interface IntegrationSettings {
  monkeyapi_url: string
  monkeyapi_key: string
  monkeyapi_enabled: boolean
  paypinga_url: string
  paypinga_key: string
  paypinga_enabled: boolean
}

export interface Settings {
  general: GeneralSettings
  integrations: IntegrationSettings
}

export function useSettings() {
  return useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: () => api.get('/settings').then(r => r.data),
  })
}

export function useUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: Settings) => api.put<Settings>('/settings', data).then(r => r.data),
    onSuccess: data => qc.setQueryData(['settings'], data),
  })
}

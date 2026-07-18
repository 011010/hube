import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { http } from '../services/api'

export interface GeneralSettings {
  display_name: string
  view_preferences?: string
}

export interface IntegrationSettings {
  monkeyapi_url: string
  monkeyapi_key_set: boolean
  monkeyapi_enabled: boolean
  paypinga_url: string
  paypinga_key_set: boolean
  paypinga_enabled: boolean
}

export interface Settings {
  general: GeneralSettings
  integrations: IntegrationSettings
}

// PUT payload: the server never returns key values, so keys are only sent
// when the user typed a replacement. An omitted (or empty) key means
// "leave unchanged".
export interface SettingsUpdatePayload {
  general: GeneralSettings
  integrations: {
    monkeyapi_url: string
    monkeyapi_enabled: boolean
    paypinga_url: string
    paypinga_enabled: boolean
    monkeyapi_key?: string
    paypinga_key?: string
  }
}

export function useSettings() {
  return useQuery<Settings>({
    queryKey: ['settings'],
    queryFn: () => http.get('/settings').then(r => r.data),
  })
}

export function useUpdateSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: SettingsUpdatePayload) => http.put<Settings>('/settings', data).then(r => r.data),
    onSuccess: data => qc.setQueryData(['settings'], data),
  })
}

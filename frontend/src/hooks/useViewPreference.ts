import { useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useSettings, useUpdateSettings, type Settings } from './useSettings'

export type PageViewKey = 'tasks_view' | 'projects_view' | 'notes_view'
export type ViewMode = 'kanban' | 'table'

export const DEFAULT_VIEW: ViewMode = 'kanban'
const VIEW_MODES: ViewMode[] = ['kanban', 'table']

function parsePrefs(raw?: string): Partial<Record<PageViewKey, unknown>> {
  try {
    const parsed = JSON.parse(raw ?? '{}')
    return parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)
      ? parsed
      : {}
  } catch {
    return {}
  }
}

function normalizeView(value: unknown): ViewMode {
  return VIEW_MODES.includes(value as ViewMode) ? (value as ViewMode) : DEFAULT_VIEW
}

export function useViewPreference(key: PageViewKey) {
  const { data } = useSettings()
  const update = useUpdateSettings()
  const queryClient = useQueryClient()

  const prefs = parsePrefs(data?.general.view_preferences)
  const value = normalizeView(prefs[key])

  // Read the latest cached settings to avoid stale closures within a single
  // component instance. Note: rapid toggles across different preference keys
  // can still race because both reads may happen before either PUT completes,
  // overwriting one update. This is acceptable for view preferences.
  const setValue = useCallback((v: ViewMode) => {
    const latest = queryClient.getQueryData<Settings>(['settings'])
    if (!latest) return
    const latestPrefs = parsePrefs(latest.general.view_preferences)
    const next = { ...latestPrefs, [key]: v }
    update.mutate({
      ...latest,
      general: { ...latest.general, view_preferences: JSON.stringify(next) },
    })
  }, [key, update, queryClient])

  return [value, setValue] as const
}

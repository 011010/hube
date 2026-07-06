import { useQueryClient } from '@tanstack/react-query'
import { useSettings, useUpdateSettings, type Settings } from './useSettings'

export type PageViewKey = 'tasks_view' | 'projects_view' | 'notes_view'
export type ViewMode = 'kanban' | 'table'

export const DEFAULT_VIEW: ViewMode = 'kanban'

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
  return value === 'table' ? 'table' : DEFAULT_VIEW
}

export function useViewPreference(key: PageViewKey) {
  const { data } = useSettings()
  const update = useUpdateSettings()
  const queryClient = useQueryClient()

  const prefs = parsePrefs(data?.general.view_preferences)
  const value = normalizeView(prefs[key])

  function setValue(v: ViewMode) {
    const latest = queryClient.getQueryData<Settings>(['settings'])
    if (!latest) return
    const latestPrefs = parsePrefs(latest.general.view_preferences)
    const next = { ...latestPrefs, [key]: v }
    update.mutate({
      ...latest,
      general: { ...latest.general, view_preferences: JSON.stringify(next) },
    })
  }

  return [value, setValue] as const
}

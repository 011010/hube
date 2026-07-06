import { useSettings, useUpdateSettings } from './useSettings'

export type PageViewKey = 'tasks_view' | 'projects_view' | 'notes_view'

function normalizeView(value: unknown): 'kanban' | 'table' {
  return value === 'table' ? 'table' : 'kanban'
}

function parsePrefs(raw?: string): Record<PageViewKey, unknown> {
  try {
    const parsed = JSON.parse(raw ?? '{}')
    return parsed !== null && typeof parsed === 'object' && !Array.isArray(parsed)
      ? (parsed as Record<PageViewKey, unknown>)
      : ({} as Record<PageViewKey, unknown>)
  } catch {
    return {} as Record<PageViewKey, unknown>
  }
}

export function useViewPreference(key: PageViewKey) {
  const { data } = useSettings()
  const update = useUpdateSettings()

  const prefs = parsePrefs(data?.general.view_preferences)
  const value = normalizeView(prefs[key])

  function setValue(v: 'kanban' | 'table') {
    if (!data) return
    const next = { ...prefs, [key]: v }
    update.mutate({
      ...data,
      general: { ...data.general, view_preferences: JSON.stringify(next) },
    })
  }

  return [value, setValue] as const
}

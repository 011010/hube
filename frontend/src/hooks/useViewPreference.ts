import { useSettings, useUpdateSettings } from './useSettings'

export type PageViewKey = 'tasks_view' | 'projects_view' | 'notes_view'

function parsePrefs(raw?: string): Record<PageViewKey, 'kanban' | 'table'> {
  try {
    return JSON.parse(raw ?? '{}')
  } catch {
    return {} as Record<PageViewKey, 'kanban' | 'table'>
  }
}

export function useViewPreference(key: PageViewKey) {
  const { data } = useSettings()
  const update = useUpdateSettings()

  const prefs = parsePrefs(data?.general.view_preferences)
  const value = prefs[key] ?? 'kanban'

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

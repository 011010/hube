import { useCallback, useMemo } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useSettings, useUpdateSettings, type Settings } from './useSettings'
import { MAX_NOTE_VIEWS, parseNoteViews, type NoteView } from '../utils/noteViews'

export type { NoteView }

// Replace the view with the same id, otherwise append.
function upsert(views: NoteView[], view: NoteView): NoteView[] {
  const index = views.findIndex(v => v.id === view.id)
  if (index === -1) return [...views, view]
  const next = [...views]
  next[index] = view
  return next
}

export function useNoteViews() {
  const { data } = useSettings()
  const update = useUpdateSettings()
  const queryClient = useQueryClient()

  const views = useMemo(() => parseNoteViews(data?.general.notes_views), [data?.general.notes_views])

  // The settings PUT is a full replacement: read the latest cached settings
  // and merge so saving a view never erases view_preferences or other fields.
  const saveViews = useCallback((next: NoteView[]) => {
    const latest = queryClient.getQueryData<Settings>(['settings'])
    if (!latest) return
    update.mutate({
      ...latest,
      general: {
        ...latest.general,
        notes_views: JSON.stringify(next.slice(0, MAX_NOTE_VIEWS)),
      },
    })
  }, [queryClient, update])

  return {
    views,
    saveView: (view: NoteView) => saveViews(upsert(views, view)),
    deleteView: (id: string) => saveViews(views.filter(v => v.id !== id)),
  }
}

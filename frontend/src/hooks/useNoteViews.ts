import { useCallback, useMemo } from 'react'
import { useQueryClient, type QueryClient } from '@tanstack/react-query'
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

// The views as they currently exist in the settings cache; null when
// settings have not loaded yet. Module-level so the hook's callbacks can
// stay referentially stable.
function readLatestViews(queryClient: QueryClient): { latest: Settings; current: NoteView[] } | null {
  const latest = queryClient.getQueryData<Settings>(['settings'])
  if (!latest) return null
  return { latest, current: parseNoteViews(latest.general.notes_views) }
}

export function useNoteViews() {
  const { data } = useSettings()
  const { mutate, isPending, error } = useUpdateSettings()
  const queryClient = useQueryClient()

  const views = useMemo(() => parseNoteViews(data?.general.notes_views), [data?.general.notes_views])

  // The settings PUT is a full replacement, so every write merges into the
  // latest cached settings instead of clobbering view_preferences or other
  // fields.
  const writeViews = useCallback((latest: Settings, next: NoteView[]) => {
    mutate({
      ...latest,
      general: {
        ...latest.general,
        notes_views: JSON.stringify(next.slice(0, MAX_NOTE_VIEWS)),
      },
    })
  }, [mutate])

  // Read the freshest cached settings on each operation so a second save in
  // the same render pass builds on the last completed write, not on this
  // render's snapshot. Note: two PUTs still in flight can overwrite each
  // other, and edits from another tab race the same way. This is acceptable
  // for saved views.
  const saveView = useCallback((view: NoteView) => {
    const latestViews = readLatestViews(queryClient)
    if (!latestViews) return
    writeViews(latestViews.latest, upsert(latestViews.current, view))
  }, [queryClient, writeViews])

  const deleteView = useCallback((id: string) => {
    const latestViews = readLatestViews(queryClient)
    if (!latestViews) return
    const next = latestViews.current.filter(v => v.id !== id)
    // Deleting a non-existent id: skip the identical-payload write.
    if (next.length === latestViews.current.length) return
    writeViews(latestViews.latest, next)
  }, [queryClient, writeViews])

  // Surfaced so a failed view save is not silent; the error clears by itself
  // when the next save starts (react-query resets the mutation on mutate).
  return { views, saveView, deleteView, saving: isPending, saveError: error }
}

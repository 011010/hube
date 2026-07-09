import { useCallback } from 'react'

/**
 * Shared kanban drag-move handler for boards keyed on a `status` field:
 * validates the target column, looks up the dragged item, no-ops when the
 * status didn't actually change, then delegates the update to the caller.
 */
export function useStatusKanbanMove<T extends { id: string; status: S }, S extends string>(
  items: T[] | undefined,
  validStatuses: readonly S[],
  updateFn: (id: string, data: { status: S }) => void,
) {
  return useCallback(
    (itemId: string, _sourceColumnId: string, targetColumnId: string) => {
      if (!validStatuses.includes(targetColumnId as S)) return
      const item = items?.find(i => i.id === itemId)
      if (!item || item.status === targetColumnId) return
      updateFn(itemId, { status: targetColumnId as S })
    },
    [items, validStatuses, updateFn],
  )
}

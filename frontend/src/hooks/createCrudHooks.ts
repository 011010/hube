import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

export interface CrudApi<T, TCreate, TUpdate> {
  list: () => Promise<T[]>
  create: (data: TCreate) => Promise<T>
  update: (id: string, data: TUpdate) => Promise<T>
  delete: (id: string) => Promise<unknown>
}

export interface CrudHooksOptions {
  /** Extra query key to invalidate on update, e.g. a single-item cache entry. */
  itemQueryKey?: (id: string) => unknown[]
}

/**
 * Factory for the common list/create/update/delete shape shared by most
 * resource hooks (tasks, events, wishlist, diagrams, notes, projects).
 * Resources with extra list params or single-item fetches keep those
 * hand-written and only reuse the create/update/delete hooks from here.
 */
export function createCrudHooks<T extends { id: string }, TCreate, TUpdate>(
  queryKey: string,
  api: CrudApi<T, TCreate, TUpdate>,
  options: CrudHooksOptions = {},
) {
  function useList() {
    return useQuery<T[]>({ queryKey: [queryKey], queryFn: api.list })
  }

  function useCreate() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: api.create,
      onSuccess: () => qc.invalidateQueries({ queryKey: [queryKey] }),
    })
  }

  function useUpdate() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: ({ id, data }: { id: string; data: TUpdate }) => api.update(id, data),
      onSuccess: (_, { id }) => {
        qc.invalidateQueries({ queryKey: [queryKey] })
        if (options.itemQueryKey) qc.invalidateQueries({ queryKey: options.itemQueryKey(id) })
      },
    })
  }

  function useDelete() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: api.delete,
      onSuccess: () => qc.invalidateQueries({ queryKey: [queryKey] }),
    })
  }

  return { useList, useCreate, useUpdate, useDelete }
}

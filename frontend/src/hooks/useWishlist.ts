import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { wishlistApi } from '../services/api'
import type { WishlistItem } from '../types'

export function useWishlist() {
  return useQuery({ queryKey: ['wishlist'], queryFn: wishlistApi.list })
}

export function useCreateWishlistItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: wishlistApi.create,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['wishlist'] }),
  })
}

export function useUpdateWishlistItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<WishlistItem> }) =>
      wishlistApi.update(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['wishlist'] }),
  })
}

export function useDeleteWishlistItem() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: wishlistApi.delete,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['wishlist'] }),
  })
}

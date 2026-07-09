import { createCrudHooks } from './createCrudHooks'
import { wishlistApi } from '../services/api'
import type { WishlistItem } from '../types'

const wishlistCrud = createCrudHooks<
  WishlistItem,
  Omit<WishlistItem, 'id' | 'created_at' | 'updated_at'>,
  Partial<WishlistItem>
>('wishlist', wishlistApi)

export const useWishlist = wishlistCrud.useList
export const useCreateWishlistItem = wishlistCrud.useCreate
export const useUpdateWishlistItem = wishlistCrud.useUpdate
export const useDeleteWishlistItem = wishlistCrud.useDelete

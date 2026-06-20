import { useState } from 'react'
import { useWishlist, useCreateWishlistItem, useUpdateWishlistItem, useDeleteWishlistItem } from '../../hooks/useWishlist'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import type { Priority, WishlistItem, WishlistStatus } from '../../types'

type ItemForm = Omit<WishlistItem, 'id' | 'created_at' | 'updated_at'>

const emptyForm = (): ItemForm => ({
  name: '',
  description: '',
  url: '',
  store: '',
  target_price: 0,
  current_price: 0,
  currency: 'USD',
  priority: 'medium',
  status: 'pending',
  notes: '',
})

function priorityVariant(p: Priority): 'default' | 'warning' | 'danger' {
  return p === 'high' ? 'danger' : p === 'medium' ? 'warning' : 'default'
}

interface ModalProps {
  open: boolean
  item?: WishlistItem
  isPending: boolean
  onClose: () => void
  onSave: (data: ItemForm) => void
}

function WishlistModal({ open, item, isPending, onClose, onSave }: ModalProps) {
  const [form, setForm] = useState<ItemForm>(item ? {
    name: item.name,
    description: item.description,
    url: item.url,
    store: item.store,
    target_price: item.target_price,
    current_price: item.current_price,
    currency: item.currency,
    priority: item.priority,
    status: item.status,
    notes: item.notes,
  } : emptyForm())

  if (!open) return null

  const set = <K extends keyof ItemForm>(k: K, v: ItemForm[K]) =>
    setForm(f => ({ ...f, [k]: v }))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="bg-gray-900 border border-gray-700 rounded-xl p-6 w-full max-w-md space-y-4">
        <h2 className="text-lg font-semibold text-white">{item ? 'Edit item' : 'New wishlist item'}</h2>

        <div className="space-y-3">
          <input
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
            placeholder="Name *"
            value={form.name}
            onChange={e => set('name', e.target.value)}
          />
          <div className="grid grid-cols-2 gap-3">
            <input
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              placeholder="Store"
              value={form.store}
              onChange={e => set('store', e.target.value)}
            />
            <input
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              placeholder="URL"
              value={form.url}
              onChange={e => set('url', e.target.value)}
            />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <input
              type="number"
              min={0}
              step={0.01}
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              placeholder="Target price"
              value={form.target_price || ''}
              onChange={e => set('target_price', parseFloat(e.target.value) || 0)}
            />
            <input
              type="number"
              min={0}
              step={0.01}
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              placeholder="Current price"
              value={form.current_price || ''}
              onChange={e => set('current_price', parseFloat(e.target.value) || 0)}
            />
            <input
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
              placeholder="Currency"
              value={form.currency}
              onChange={e => set('currency', e.target.value.toUpperCase())}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <select
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
              value={form.priority}
              onChange={e => set('priority', e.target.value as Priority)}
            >
              <option value="low">Low priority</option>
              <option value="medium">Medium priority</option>
              <option value="high">High priority</option>
            </select>
            <select
              className="bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
              value={form.status}
              onChange={e => set('status', e.target.value as WishlistStatus)}
            >
              <option value="pending">Pending</option>
              <option value="purchased">Purchased</option>
            </select>
          </div>
          <textarea
            rows={2}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500 resize-none"
            placeholder="Notes"
            value={form.notes}
            onChange={e => set('notes', e.target.value)}
          />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button variant="ghost" onClick={onClose}>Cancel</Button>
          <Button onClick={() => onSave(form)} disabled={!form.name || isPending}>
            {isPending ? 'Saving…' : 'Save'}
          </Button>
        </div>
      </div>
    </div>
  )
}

export function WishlistPage() {
  const { data: items = [], isLoading } = useWishlist()
  const create = useCreateWishlistItem()
  const update = useUpdateWishlistItem()
  const remove = useDeleteWishlistItem()

  const [createOpen, setCreateOpen] = useState(false)
  const [editItem, setEditItem] = useState<WishlistItem | null>(null)

  const handleCreate = (data: ItemForm) => {
    create.mutate(data, { onSuccess: () => setCreateOpen(false) })
  }

  const handleUpdate = (data: ItemForm) => {
    if (!editItem) return
    update.mutate({ id: editItem.id, data: { ...editItem, ...data } }, {
      onSuccess: () => setEditItem(null),
    })
  }

  const togglePurchased = (item: WishlistItem) => {
    const status: WishlistStatus = item.status === 'purchased' ? 'pending' : 'purchased'
    update.mutate({ id: item.id, data: { ...item, status } })
  }

  const pending = items.filter(i => i.status === 'pending')
  const purchased = items.filter(i => i.status === 'purchased')

  const priceLabel = (item: WishlistItem) => {
    if (item.target_price === 0) return null
    const curr = item.currency || 'USD'
    const below = item.current_price > 0 && item.current_price <= item.target_price
    return (
      <span className={`text-xs ${below ? 'text-emerald-400' : 'text-gray-500'}`}>
        {item.current_price > 0 ? `${curr} ${item.current_price.toFixed(2)} / ` : ''}
        target {curr} {item.target_price.toFixed(2)}
        {below && ' ✓'}
      </span>
    )
  }

  const renderGroup = (group: WishlistItem[], label: string) => {
    if (group.length === 0) return null
    return (
      <section>
        <h2 className="text-xs font-semibold text-gray-500 uppercase tracking-widest mb-2">
          {label} · {group.length}
        </h2>
        <ul className="space-y-2">
          {group.map(item => (
            <li
              key={item.id}
              className="flex items-center gap-3 bg-gray-900 border border-gray-800 rounded-lg px-4 py-3 group"
            >
              <input
                type="checkbox"
                checked={item.status === 'purchased'}
                onChange={() => togglePurchased(item)}
                className="w-4 h-4 accent-indigo-500 shrink-0"
              />
              <button
                onClick={() => setEditItem(item)}
                className={`flex-1 min-w-0 text-sm text-left transition-colors ${
                  item.status === 'purchased'
                    ? 'line-through text-gray-500'
                    : 'text-gray-200 hover:text-white'
                }`}
              >
                <div className="flex items-center gap-2">
                  <span>{item.name}</span>
                  {item.store && <span className="text-xs text-gray-500">{item.store}</span>}
                </div>
                {priceLabel(item)}
                {item.notes && (
                  <span className="block text-xs text-gray-500 mt-0.5 truncate">{item.notes}</span>
                )}
              </button>
              <Badge label={item.priority} variant={priorityVariant(item.priority)} />
              {item.url && (
                <a
                  href={item.url}
                  target="_blank"
                  rel="noreferrer"
                  className="text-gray-600 hover:text-indigo-400 text-xs transition-colors"
                  onClick={e => e.stopPropagation()}
                >
                  ↗
                </a>
              )}
              <button
                onClick={() => remove.mutate(item.id)}
                className="text-gray-700 hover:text-red-400 transition-colors text-xs opacity-0 group-hover:opacity-100"
              >
                ✕
              </button>
            </li>
          ))}
        </ul>
      </section>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-white">Wishlist</h1>
        <Button onClick={() => setCreateOpen(true)}>+ New item</Button>
      </div>

      {isLoading && <p className="text-gray-400 text-sm">Loading…</p>}

      {renderGroup(pending, 'Pending')}
      {renderGroup(purchased, 'Purchased')}

      {items.length === 0 && !isLoading && (
        <p className="text-gray-500 text-sm">No items yet. Add something you want to buy.</p>
      )}

      <WishlistModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onSave={handleCreate}
        isPending={create.isPending}
      />
      <WishlistModal
        open={Boolean(editItem)}
        item={editItem ?? undefined}
        onClose={() => setEditItem(null)}
        onSave={handleUpdate}
        isPending={update.isPending}
      />
    </div>
  )
}

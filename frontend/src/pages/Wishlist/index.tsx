import { useState } from 'react'
import { safeHref } from '../../utils/url'
import { Check, Trash2, Heart, ExternalLink } from 'lucide-react'
import { useWishlist, useCreateWishlistItem, useUpdateWishlistItem, useDeleteWishlistItem } from '../../hooks/useWishlist'
import { Badge } from '../../components/atoms/Badge'
import { Button } from '../../components/atoms/Button'
import { Input } from '../../components/atoms/Input'
import { Textarea } from '../../components/atoms/Textarea'
import { Select } from '../../components/atoms/Select'
import { Modal } from '../../components/atoms/Modal'
import { PageHeader } from '../../components/molecules/PageHeader'
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

interface WishlistModalProps {
  open: boolean
  item?: WishlistItem
  isPending: boolean
  onClose: () => void
  onSave: (data: ItemForm) => void
}

function WishlistModal({ open, item, isPending, onClose, onSave }: WishlistModalProps) {
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

  const set = <K extends keyof ItemForm>(k: K, v: ItemForm[K]) =>
    setForm(f => ({ ...f, [k]: v }))

  return (
    <Modal open={open} onClose={onClose} title={item ? 'Edit item' : 'New wishlist item'}>
      <div className="space-y-3">
        <Input
          placeholder="Name *"
          value={form.name}
          onChange={e => set('name', (e.target as HTMLInputElement).value)}
        />
        <div className="grid grid-cols-2 gap-3">
          <Input
            placeholder="Store"
            value={form.store}
            onChange={e => set('store', (e.target as HTMLInputElement).value)}
          />
          <Input
            placeholder="URL"
            value={form.url}
            onChange={e => set('url', (e.target as HTMLInputElement).value)}
          />
        </div>
        <div className="grid grid-cols-3 gap-3">
          <Input
            type="number"
            min={0}
            step={0.01}
            placeholder="Target price"
            value={form.target_price || ''}
            onChange={e => set('target_price', parseFloat((e.target as HTMLInputElement).value) || 0)}
          />
          <Input
            type="number"
            min={0}
            step={0.01}
            placeholder="Current price"
            value={form.current_price || ''}
            onChange={e => set('current_price', parseFloat((e.target as HTMLInputElement).value) || 0)}
          />
          <Input
            placeholder="Currency"
            value={form.currency}
            onChange={e => set('currency', (e.target as HTMLInputElement).value.toUpperCase())}
          />
        </div>
        <div className="grid grid-cols-2 gap-3">
          <Select
            label="Priority"
            value={form.priority}
            onChange={e => set('priority', e.target.value as Priority)}
            options={[
              { value: 'low', label: 'Low priority' },
              { value: 'medium', label: 'Medium priority' },
              { value: 'high', label: 'High priority' },
            ]}
          />
          <Select
            label="Status"
            value={form.status}
            onChange={e => set('status', e.target.value as WishlistStatus)}
            options={[
              { value: 'pending', label: 'Pending' },
              { value: 'purchased', label: 'Purchased' },
            ]}
          />
        </div>
        <Textarea
          rows={2}
          placeholder="Notes"
          value={form.notes}
          onChange={e => set('notes', (e.target as HTMLTextAreaElement).value)}
        />
      </div>

      <div className="flex justify-end gap-2 pt-4">
        <Button variant="ghost" onClick={onClose}>Cancel</Button>
        <Button onClick={() => onSave(form)} disabled={!form.name || isPending}>
          {isPending ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </Modal>
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
      <span className={`text-xs ${below ? 'text-emerald-400' : 'text-text-muted'}`}>
        {item.current_price > 0 ? `${curr} ${item.current_price.toFixed(2)} / ` : ''}
        target {curr} {item.target_price.toFixed(2)}
        {below && <><Check size={12} className="inline" /> </>}
      </span>
    )
  }

  const renderGroup = (group: WishlistItem[], label: string) => {
    if (group.length === 0) return null
    return (
      <section>
        <h2 className="text-xs font-semibold text-text-muted uppercase tracking-widest mb-2">
          {label} · {group.length}
        </h2>
        <ul className="space-y-2">
          {group.map(item => (
            <li
              key={item.id}
              className="flex items-center gap-3 bg-surface-elevated border border-border rounded-lg px-4 py-3 group"
            >
              <input
                type="checkbox"
                checked={item.status === 'purchased'}
                onChange={() => togglePurchased(item)}
                className="w-4 h-4 accent-(--color-accent) shrink-0"
              />
              <button
                onClick={() => setEditItem(item)}
                className={`flex-1 min-w-0 text-sm text-left transition-colors ${
                  item.status === 'purchased'
                    ? 'line-through text-text-muted'
                    : 'text-text-secondary hover:text-text-primary'
                }`}
              >
                <div className="flex items-center gap-2">
                  <span>{item.name}</span>
                  {item.store && <span className="text-xs text-text-muted">{item.store}</span>}
                </div>
                {priceLabel(item)}
                {item.notes && (
                  <span className="block text-xs text-text-muted mt-0.5 truncate">{item.notes}</span>
                )}
              </button>
              <Badge label={item.priority} variant={priorityVariant(item.priority)} />
              {item.url && (
                <a
                  href={safeHref(item.url)}
                  target="_blank"
                  rel="noreferrer"
                  className="text-text-muted hover:text-(--color-accent) transition-colors"
                  onClick={e => e.stopPropagation()}
                >
                  <ExternalLink size={14} />
                </a>
              )}
              <button
                onClick={() => remove.mutate(item.id)}
                className="text-text-muted hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
                aria-label="Delete item"
              >
                <Trash2 size={14} />
              </button>
            </li>
          ))}
        </ul>
      </section>
    )
  }

  return (
    <div className="p-8 space-y-6">
      <PageHeader
        title="Wishlist"
        actions={
          <Button onClick={() => setCreateOpen(true)} icon={<Heart size={16} />}>New item</Button>
        }
      />

      {isLoading && <p className="text-text-muted text-sm">Loading…</p>}

      {renderGroup(pending, 'Pending')}
      {renderGroup(purchased, 'Purchased')}

      {items.length === 0 && !isLoading && (
        <p className="text-text-muted text-sm">No items yet. Add something you want to buy.</p>
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

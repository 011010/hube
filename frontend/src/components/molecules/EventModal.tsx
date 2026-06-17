import { useState, useEffect } from 'react'
import { Modal } from '../atoms/Modal'
import { Button } from '../atoms/Button'
import type { CalendarEvent } from '../../types'

const COLORS = ['#6366f1', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#f97316']

interface EventModalProps {
  open: boolean
  onClose: () => void
  onSave: (data: Omit<CalendarEvent, 'id' | 'created_at' | 'updated_at'>) => void
  onDelete?: () => void
  event?: CalendarEvent | null
  defaultDate?: string
}

function toLocalDateTimeInput(iso: string) {
  return iso ? iso.slice(0, 16) : ''
}

function toISOFromInput(local: string) {
  return local ? new Date(local).toISOString() : ''
}

export function EventModal({ open, onClose, onSave, onDelete, event, defaultDate }: EventModalProps) {
  const isEdit = !!event

  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [startAt, setStartAt] = useState('')
  const [endAt, setEndAt] = useState('')
  const [allDay, setAllDay] = useState(false)
  const [color, setColor] = useState(COLORS[0])

  useEffect(() => {
    if (event) {
      setTitle(event.title)
      setDescription(event.description)
      setStartAt(toLocalDateTimeInput(event.start_at))
      setEndAt(toLocalDateTimeInput(event.end_at))
      setAllDay(event.all_day)
      setColor(event.color)
    } else {
      setTitle('')
      setDescription('')
      setStartAt(defaultDate ? `${defaultDate}T09:00` : '')
      setEndAt(defaultDate ? `${defaultDate}T10:00` : '')
      setAllDay(false)
      setColor(COLORS[0])
    }
  }, [event, defaultDate, open])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim() || !startAt || !endAt) return
    onSave({
      title: title.trim(),
      description,
      start_at: toISOFromInput(startAt),
      end_at: toISOFromInput(endAt),
      all_day: allDay,
      color,
    })
    onClose()
  }

  return (
    <Modal open={open} onClose={onClose} title={isEdit ? 'Edit event' : 'New event'}>
      <form onSubmit={handleSubmit} className="space-y-4">
        <input
          autoFocus
          value={title}
          onChange={e => setTitle(e.target.value)}
          placeholder="Event title"
          required
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500"
        />

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            id="allday"
            checked={allDay}
            onChange={e => setAllDay(e.target.checked)}
            className="accent-indigo-500"
          />
          <label htmlFor="allday" className="text-sm text-gray-400">All day</label>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="text-xs text-gray-500 mb-1 block">Start</label>
            <input
              type={allDay ? 'date' : 'datetime-local'}
              value={allDay ? startAt.slice(0, 10) : startAt}
              onChange={e => setStartAt(allDay ? `${e.target.value}T00:00` : e.target.value)}
              required
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
            />
          </div>
          <div>
            <label className="text-xs text-gray-500 mb-1 block">End</label>
            <input
              type={allDay ? 'date' : 'datetime-local'}
              value={allDay ? endAt.slice(0, 10) : endAt}
              onChange={e => setEndAt(allDay ? `${e.target.value}T23:59` : e.target.value)}
              required
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <textarea
          value={description}
          onChange={e => setDescription(e.target.value)}
          placeholder="Description (optional)"
          rows={2}
          className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-indigo-500 resize-none"
        />

        <div>
          <label className="text-xs text-gray-500 mb-2 block">Color</label>
          <div className="flex gap-2">
            {COLORS.map(c => (
              <button
                key={c}
                type="button"
                onClick={() => setColor(c)}
                className="w-6 h-6 rounded-full transition-transform hover:scale-110"
                style={{
                  backgroundColor: c,
                  outline: color === c ? `2px solid white` : 'none',
                  outlineOffset: '2px',
                }}
              />
            ))}
          </div>
        </div>

        <div className="flex justify-between pt-1">
          {onDelete ? (
            <Button type="button" variant="danger" size="sm" onClick={onDelete}>
              Delete
            </Button>
          ) : <span />}
          <div className="flex gap-2">
            <Button type="button" variant="ghost" size="sm" onClick={onClose}>Cancel</Button>
            <Button type="submit" size="sm">Save</Button>
          </div>
        </div>
      </form>
    </Modal>
  )
}

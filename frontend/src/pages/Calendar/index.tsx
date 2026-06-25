import { useState, useMemo } from 'react'
import { ChevronLeft, ChevronRight, CalendarPlus } from 'lucide-react'
import { useEvents, useCreateEvent, useUpdateEvent, useDeleteEvent } from '../../hooks/useEvents'
import { EventChip } from '../../components/molecules/EventChip'
import { EventModal } from '../../components/molecules/EventModal'
import { Button } from '../../components/atoms/Button'
import type { CalendarEvent } from '../../types'

const DAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const MONTHS = ['January','February','March','April','May','June','July','August','September','October','November','December']

function toYMD(date: Date) {
  return date.toISOString().slice(0, 10)
}

function buildGrid(year: number, month: number): Date[] {
  const first = new Date(year, month, 1)
  const last = new Date(year, month + 1, 0)
  const cells: Date[] = []

  for (let i = 0; i < first.getDay(); i++) {
    cells.push(new Date(year, month, -first.getDay() + i + 1))
  }
  for (let d = 1; d <= last.getDate(); d++) {
    cells.push(new Date(year, month, d))
  }
  const remaining = 42 - cells.length
  for (let i = 1; i <= remaining; i++) {
    cells.push(new Date(year, month + 1, i))
  }
  return cells
}

export function CalendarPage() {
  const today = new Date()
  const [cursor, setCursor] = useState({ year: today.getFullYear(), month: today.getMonth() })
  const [modalOpen, setModalOpen] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null)
  const [selectedDate, setSelectedDate] = useState<string>('')

  const rangeFrom = new Date(cursor.year, cursor.month, 1).toISOString()
  const rangeTo = new Date(cursor.year, cursor.month + 1, 0, 23, 59, 59).toISOString()

  const { data: events = [] } = useEvents(rangeFrom, rangeTo)
  const createEvent = useCreateEvent()
  const updateEvent = useUpdateEvent()
  const deleteEvent = useDeleteEvent()

  const grid = useMemo(() => buildGrid(cursor.year, cursor.month), [cursor])

  const eventsByDay = useMemo(() => {
    const map: Record<string, CalendarEvent[]> = {}
    for (const ev of events) {
      const key = ev.start_at.slice(0, 10)
      if (!map[key]) map[key] = []
      map[key].push(ev)
    }
    return map
  }, [events])

  const prevMonth = () =>
    setCursor(c => c.month === 0 ? { year: c.year - 1, month: 11 } : { ...c, month: c.month - 1 })

  const nextMonth = () =>
    setCursor(c => c.month === 11 ? { year: c.year + 1, month: 0 } : { ...c, month: c.month + 1 })

  const openNew = (date: string) => {
    setSelectedEvent(null)
    setSelectedDate(date)
    setModalOpen(true)
  }

  const openEdit = (ev: CalendarEvent) => {
    setSelectedEvent(ev)
    setSelectedDate('')
    setModalOpen(true)
  }

  const handleSave = (data: Omit<CalendarEvent, 'id' | 'created_at' | 'updated_at'>) => {
    if (selectedEvent) {
      updateEvent.mutate({ id: selectedEvent.id, data: { ...selectedEvent, ...data } })
    } else {
      createEvent.mutate(data)
    }
  }

  const handleDelete = () => {
    if (!selectedEvent) return
    deleteEvent.mutate(selectedEvent.id)
    setModalOpen(false)
  }

  const todayStr = toYMD(today)
  const currentMonth = cursor.month

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center justify-between px-8 py-5 border-b border-border bg-surface-elevated">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-semibold text-text-primary">
            {MONTHS[cursor.month]} {cursor.year}
          </h1>
          <div className="flex gap-1">
            <button
              onClick={prevMonth}
              className="w-8 h-8 flex items-center justify-center rounded-lg text-text-muted hover:bg-surface-card hover:text-text-primary transition-colors"
              aria-label="Previous month"
            >
              <ChevronLeft size={16} />
            </button>
            <button
              onClick={nextMonth}
              className="w-8 h-8 flex items-center justify-center rounded-lg text-text-muted hover:bg-surface-card hover:text-text-primary transition-colors"
              aria-label="Next month"
            >
              <ChevronRight size={16} />
            </button>
          </div>
          <button
            onClick={() => setCursor({ year: today.getFullYear(), month: today.getMonth() })}
            className="text-xs text-text-muted hover:text-text-primary transition-colors px-2 py-1 rounded-md hover:bg-surface-card"
          >
            Today
          </button>
        </div>
        <Button size="sm" icon={<CalendarPlus size={14} />} onClick={() => openNew(todayStr)}>Event</Button>
      </div>

      {/* Day headers */}
      <div className="grid grid-cols-7 border-b border-border">
        {DAYS.map(d => (
          <div key={d} className="py-2 text-center text-xs font-medium text-text-muted uppercase tracking-wider">
            {d}
          </div>
        ))}
      </div>

      {/* Grid */}
      <div className="flex-1 grid grid-cols-7 grid-rows-6 overflow-hidden">
        {grid.map((date, i) => {
          const ymd = toYMD(date)
          const isCurrentMonth = date.getMonth() === currentMonth
          const isToday = ymd === todayStr
          const dayEvents = eventsByDay[ymd] || []

          return (
            <div
              key={i}
              onClick={() => openNew(ymd)}
              className={`border-b border-r border-border p-1.5 min-h-0 cursor-pointer group transition-colors
                ${isCurrentMonth ? 'bg-surface-base hover:bg-surface-elevated' : 'bg-surface-base/50 hover:bg-surface-elevated/50'}
                ${i % 7 === 0 ? 'border-l-0' : ''}
              `}
            >
              <div className="flex items-center justify-between mb-1">
                <span
                  className={`text-xs font-medium w-6 h-6 flex items-center justify-center rounded-full
                    ${isToday ? 'bg-(--color-accent) text-white' : isCurrentMonth ? 'text-text-secondary' : 'text-text-muted'}
                  `}
                >
                  {date.getDate()}
                </span>
                {dayEvents.length > 0 && (
                  <span className="text-xs text-text-muted group-hover:text-text-secondary opacity-0 group-hover:opacity-100 transition-opacity">
                    +
                  </span>
                )}
              </div>
              <div className="space-y-0.5 overflow-hidden">
                {dayEvents.slice(0, 3).map(ev => (
                  <EventChip
                    key={ev.id}
                    event={ev}
                    onClick={e => { e.stopPropagation(); openEdit(ev) }}
                  />
                ))}
                {dayEvents.length > 3 && (
                  <span className="text-xs text-text-muted pl-1">+{dayEvents.length - 3} more</span>
                )}
              </div>
            </div>
          )
        })}
      </div>

      <EventModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSave={handleSave}
        onDelete={selectedEvent ? handleDelete : undefined}
        event={selectedEvent}
        defaultDate={selectedDate}
      />
    </div>
  )
}

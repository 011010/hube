import type { CalendarEvent } from '../../types'

interface EventChipProps {
  event: CalendarEvent
  onClick: (e: React.MouseEvent) => void
}

export function EventChip({ event, onClick }: EventChipProps) {
  return (
    <button
      onClick={onClick}
      className="w-full text-left px-1.5 py-0.5 rounded text-xs font-medium truncate leading-5 transition-opacity hover:opacity-80"
      style={{ backgroundColor: event.color + '33', color: event.color }}
      title={event.title}
    >
      {event.title}
    </button>
  )
}

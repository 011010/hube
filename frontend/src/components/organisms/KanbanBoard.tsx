import { useMemo, useState } from 'react'
import type { ReactNode } from 'react'
import { GripVertical } from 'lucide-react'
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  useDroppable,
  rectIntersection,
} from '@dnd-kit/core'
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core'
import {
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'

export interface KanbanColumn<T> {
  id: string
  title: string
  items: T[]
}

export interface KanbanBoardProps<T> {
  columns: KanbanColumn<T>[]
  renderCard: (item: T) => ReactNode
  getItemId: (item: T) => string
  onMove: (itemId: string, sourceColumnId: string, targetColumnId: string) => void
}

function uniqueCardId(columnId: string, itemId: string): string {
  return `${columnId}::${itemId}`
}

const cardSurfaceClass =
  'rounded-lg border border-border-subtle bg-surface-card'

interface CardMeta<T> {
  columnId: string
  itemId: string
  item: T
}

function useCardMeta<T>(
  columns: KanbanColumn<T>[],
  getItemId: (item: T) => string,
) {
  return useMemo(() => {
    const byUniqueId = new Map<string, CardMeta<T>>()
    const columnByUniqueId = new Map<string, string>()

    for (const column of columns) {
      for (const item of column.items) {
        const itemId = getItemId(item)
        const uniqueId = uniqueCardId(column.id, itemId)
        byUniqueId.set(uniqueId, { columnId: column.id, itemId, item })
        columnByUniqueId.set(uniqueId, column.id)
      }
    }

    return { byUniqueId, columnByUniqueId }
  }, [columns, getItemId])
}

interface SortableCardProps<T> {
  item: T
  columnId: string
  getItemId: (item: T) => string
  renderCard: (item: T) => ReactNode
}

function SortableCard<T>({
  item,
  columnId,
  getItemId,
  renderCard,
}: SortableCardProps<T>) {
  const itemId = getItemId(item)
  const uniqueId = uniqueCardId(columnId, itemId)
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: uniqueId,
    data: { columnId, itemId },
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={[
        cardSurfaceClass,
        'group select-none transition-opacity',
        isDragging ? 'opacity-40' : 'opacity-100',
      ].join(' ')}
    >
      <div
        {...attributes}
        {...listeners}
        className="cursor-grab active:cursor-grabbing py-1 flex justify-center text-text-muted hover:text-text-primary touch-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-(--color-accent) focus-visible:rounded"
      >
        <GripVertical size={16} />
      </div>
      <div className="px-3 pb-3">
        {renderCard(item)}
      </div>
    </div>
  )
}

interface KanbanColumnProps<T> {
  column: KanbanColumn<T>
  getItemId: (item: T) => string
  renderCard: (item: T) => ReactNode
}

function KanbanColumnComponent<T>({
  column,
  getItemId,
  renderCard,
}: KanbanColumnProps<T>) {
  const { setNodeRef, isOver } = useDroppable({
    id: column.id,
    data: { columnId: column.id },
  })

  const itemIds = useMemo(
    () => column.items.map((item) => uniqueCardId(column.id, getItemId(item))),
    [column, getItemId],
  )

  return (
    <div
      ref={setNodeRef}
      className={[
        'flex w-80 shrink-0 flex-col rounded-xl border border-border',
        'bg-surface-elevated overflow-hidden',
        isOver ? 'ring-2 ring-(--color-accent)' : '',
      ].join(' ')}
    >
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <h3 className="font-semibold text-text-primary">{column.title}</h3>
        <span className="text-sm text-text-muted">{column.items.length}</span>
      </div>

      <div className="min-h-[120px] flex-1 p-3">
        <SortableContext
          items={itemIds}
          strategy={verticalListSortingStrategy}
        >
          <div className="flex flex-col gap-2">
            {column.items.map((item) => (
              <SortableCard
                key={uniqueCardId(column.id, getItemId(item))}
                item={item}
                columnId={column.id}
                getItemId={getItemId}
                renderCard={renderCard}
              />
            ))}
          </div>
        </SortableContext>
      </div>
    </div>
  )
}

export function KanbanBoard<T>({
  columns,
  renderCard,
  getItemId,
  onMove,
}: KanbanBoardProps<T>) {
  // Keyboard navigation can reorder cards within a column only.
  // Cross-column keyboard moves are not supported yet.
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  )

  const [activeUniqueId, setActiveUniqueId] = useState<string | null>(null)
  const { byUniqueId, columnByUniqueId } = useCardMeta(columns, getItemId)

  const activeMeta = activeUniqueId
    ? byUniqueId.get(activeUniqueId) ?? null
    : null

  const handleDragStart = (event: DragStartEvent) => {
    setActiveUniqueId(String(event.active.id))
  }

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    setActiveUniqueId(null)

    if (!over) return

    const source = byUniqueId.get(String(active.id))
    if (!source) return

    const overId = String(over.id)
    const targetColumnId =
      columns.find((column) => column.id === overId)?.id ??
      columnByUniqueId.get(overId)

    // Same-column drops are ignored; intra-column reordering is not supported.
    if (targetColumnId && targetColumnId !== source.columnId) {
      onMove(source.itemId, source.columnId, targetColumnId)
    }
  }

  const handleDragCancel = () => {
    setActiveUniqueId(null)
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={rectIntersection}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      <div className="flex gap-4 overflow-x-auto pb-2">
        {columns.map((column) => (
          <KanbanColumnComponent
            key={column.id}
            column={column}
            getItemId={getItemId}
            renderCard={renderCard}
          />
        ))}
      </div>

      <DragOverlay>
        {activeMeta ? (
          <div className={`${cardSurfaceClass} p-3 rotate-2 opacity-90`}>
            {renderCard(activeMeta.item)}
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
